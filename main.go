package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"ocr.service.worker/config"
	"ocr.service.worker/model"
	"ocr.service.worker/module"
	"os"
	"time"
)

type Worker struct {
	client            *http.Client
	rabbitmq          *module.RabbitMQ
	imageSuccessQueue string
	imageErrorQueue   string
	imageTaskQueue    string
	aiUrl             string
}

func (q *Worker) CallAICore(image model.Image) ([]byte, error) {
	strImage, _ := json.Marshal(image)
	req, err := http.NewRequest(http.MethodGet, q.aiUrl, bytes.NewReader(strImage))
	if err != nil {
		return nil, err
	}
	res, err := q.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(body))
	return body, nil
}
func (q *Worker) HandleTask(message []byte, messageAction *module.MessageAction) {
	var image model.Image
	err := json.Unmarshal(message, &image)
	if err != nil {
		fmt.Println(err)
		messageAction.Ack()
		return
	}
	fmt.Println("[RECEIVE]: ", image)
	time.Sleep(3 * time.Second)
	data, err := q.CallAICore(image)
	if err != nil {
		image.Status = "error"
		image.Error = err.Error()
		imageResponse, _ := json.Marshal(image)
		fmt.Println("[SEND] error: ", string(imageResponse))
		err = q.rabbitmq.SendMessage(q.imageErrorQueue, imageResponse, 0)
	} else {
		image.Status = "done"
		image.Data = string(data)
		imageResponse, _ := json.Marshal(image)
		fmt.Println("[SEND] success: ", imageResponse)
		err = q.rabbitmq.SendMessage(q.imageSuccessQueue, imageResponse, 0)
	}
	if err != nil {
		messageAction.Reject()
	} else {
		messageAction.Ack()
	}
}

func NewWorker() (*Worker, error) {
	CONFIG, _ := config.NewConfig(nil)
	var q Worker
	var err error
	q.imageTaskQueue = CONFIG.GetString("IMAGE_TASK_QUEUE")
	q.imageSuccessQueue = CONFIG.GetString("IMAGE_SUCCESS_QUEUE")
	q.imageErrorQueue = CONFIG.GetString("IMAGE_ERROR_QUEUE")
	q.aiUrl = CONFIG.GetString("AI_URL")
	q.client = module.CreateClient()
	rabbitmqLogin := module.RabbitMQLogin{
		Host:     CONFIG.GetString("RABBITMQ_HOST"),
		Port:     CONFIG.GetString("RABBITMQ_PORT"),
		Username: CONFIG.GetString("RABBITMQ_USERNAME"),
		Password: CONFIG.GetString("RABBITMQ_PASSWORD"),
		VHOST:    CONFIG.GetString("RABBITMQ_VHOST"),
	}
	q.rabbitmq, err = module.NewRabbitMQ(rabbitmqLogin)
	if err != nil {
		return nil, err
	}
	err = q.rabbitmq.CreateQueue(q.imageTaskQueue, 10)
	if err != nil {
		return nil, err
	}
	err = q.rabbitmq.CreateQueue(q.imageSuccessQueue, 0)
	if err != nil {
		return nil, err
	}
	err = q.rabbitmq.CreateQueue(q.imageErrorQueue, 0)
	if err != nil {
		return nil, err
	}
	var consumeProcessRequest = module.Consume{
		q.imageTaskQueue,
		"",
		false,
		false,
		false,
		false,
		10,
	}
	q.rabbitmq.Consume(consumeProcessRequest, q.HandleTask)
	return &q, err
}

func main() {
	forever := make(chan bool)
	fmt.Println("service start")
	_, err := NewWorker()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	<-forever
}
