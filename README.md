## Description
dịch vụ này sẽ nhận task từ rabbitmq, gọi AI core và trả kết quả thông qua rabbitmq 

## REQUIREMENT
```
- Rabbitmq
```

## INSTALL
```bash
#install golang
sudo snap install go --classic
```

## RUN
```bash
go run .
```

## CONFIG
```.env
RABBITMQ_HOST=host
RABBITMQ_PORT=5672
RABBITMQ_USERNAME=
RABBITMQ_PASSWORD=
RABBITMQ_VHOST=/
```

## Docker
```bash
docker build -t ocr.service.worker:1.0.0 .
```

## TEST (not yet)
unit test
```bash
go test $(go list ./... | grep -v /vendor/ | grep -v /test)
```
e2e test
```bash
go test ./test
```