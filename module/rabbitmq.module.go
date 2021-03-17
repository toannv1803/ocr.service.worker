package module

import (
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"os"
	"sync"
	"time"
)

type RabbitMQLogin struct {
	Username string
	Password string
	Host     string
	Port     string
	VHOST    string
}

type Consume struct {
	Queue     string
	Consumer  string
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Prefetch  int
}

type MessageAction struct {
	Ack    func() error
	Reject func() error
}

func NewRabbitMQ(amqpLogin RabbitMQLogin) (*RabbitMQ, error) {
	var p RabbitMQ
	p.amqpLogin = amqpLogin
	p.arrConsume = make([]Consume, 0)
	p.arrHandleMessage = make([]func(message []byte, messageAction *MessageAction), 0)
	p.err = make(chan error)
	err := p.Connect()
	if err != nil {
		return nil, err
	}
	go p.IntervalCheckConnectionClose()
	return &p, nil
}

type RabbitMQ struct {
	ch               *amqp.Channel
	err              chan error
	amqpLogin        RabbitMQLogin
	mux              sync.Mutex
	arrConsume       []Consume
	arrHandleMessage []func(message []byte, messageAction *MessageAction)
	logger           *logrus.Logger
}

func (q *RabbitMQ) SetLogger(logger *logrus.Logger) {
	q.logger = logger
}

func (q *RabbitMQ) Connect() error {
	if q.logger != nil {
		q.logger.Info("rbmq connecting")
	}
	amqpLogin := q.amqpLogin
	if amqpLogin == (RabbitMQLogin{}) {
		amqpLogin.Username = "guest"
		amqpLogin.Password = "guest"
		amqpLogin.Host = "localhost"
		amqpLogin.Port = "5672"
		amqpLogin.VHOST = "/"
	}
	conn, err := amqp.Dial("amqp://" + amqpLogin.Username + ":" + amqpLogin.Password + "@" + amqpLogin.Host + ":" + amqpLogin.Port + "/" + amqpLogin.VHOST)
	if err != nil {
		return err
	}
	go func() {
		<-conn.NotifyClose(make(chan *amqp.Error)) //Listen to NotifyClose
		if q.logger != nil {
			q.logger.Warn("rbmq connection closed")
		}
		q.err <- errors.New("connection closed")
	}()
	q.ch, err = conn.Channel()
	if err != nil {
		return err
	}
	if len(q.arrConsume) != 0 {
		for i := range q.arrConsume {
			q.Consume(q.arrConsume[i], q.arrHandleMessage[i])
		}
	}
	go func() {
		<-q.ch.NotifyClose(make(chan *amqp.Error)) //Listen to NotifyClose
		if q.logger != nil {
			q.logger.Warn("rbmq channel closed")
		}
		q.err <- errors.New("channel closed")
	}()
	return nil
}

func (q *RabbitMQ) Reconnect() error {
	var err error
	for i := 0; i < 6; i++ {
		err = q.Connect()
		if err == nil {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		os.Exit(1)
	}
	return nil
}

func (q *RabbitMQ) IntervalCheckConnectionClose() {
	for {
		<-q.err
		q.Reconnect()
	}
}

func (q *RabbitMQ) Consume(consume Consume, handleMessage func(message []byte, messageAction *MessageAction)) error {
	q.arrConsume = append(q.arrConsume, consume)
	q.arrHandleMessage = append(q.arrHandleMessage, handleMessage)
	err := q.ch.Qos(
		consume.Prefetch, // prefetch count
		0,                // prefetch size
		false,            // global
	)
	if err != nil {
		return err
	}
	msgs, err := q.ch.Consume(
		consume.Queue,
		consume.Consumer,
		consume.AutoAck,
		consume.Exclusive,
		consume.NoLocal,
		consume.NoWait,
		nil,
	)
	if err != nil {
		return err
	}
	if handleMessage == nil {
		return errors.New("require handleMessage")
	}
	go func() {
		for d := range msgs {
			var Ack = func() error {
				err := d.Ack(false)
				if q.logger != nil && err != nil {
					q.logger.WithError(err).Info("rbmq ack error")
				}
				return err
			}
			var Reject = func() error {
				err := d.Reject(true)
				if q.logger != nil && err != nil {
					q.logger.WithError(err).Info("rbmq reject error")
				}
				return err
			}
			handleMessage(d.Body, &MessageAction{Ack, Reject})
		}
	}()
	return nil
}

func (q *RabbitMQ) SendMessage(queue string, data []byte, priority int) error {
	if queue == "" {
		return errors.New("queue is null")
	}
	select { //non blocking channel - if there is no error will go to default where we do nothing
	case err := <-q.err:
		if err != nil {
			q.Reconnect()
		}
	default:
	}
	err := q.ch.Publish(
		"",    // exchange
		queue, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        data,
			Priority:    uint8(priority),
		})
	return err
}

func (q *RabbitMQ) CreateQueue(queue string, priority int64) error {
	args := make(amqp.Table)
	args["x-max-priority"] = priority
	var err error
	if priority > 0 {
		_, err = q.ch.QueueDeclare(
			queue, // name
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			args,  // arguments
		)
	} else {
		_, err = q.ch.QueueDeclare(
			queue, // name
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
	}
	if err != nil {
		return errors.New("[CreateQueue]" + queue + err.Error())
	}

	return nil
}
