FROM golang:alpine3.12 AS builder
WORKDIR /go/src/app
COPY . .
RUN go get -d -v
RUN go build -o /go/bin/ocr.service.worker

FROM alpine:3.12
COPY --from=builder /go/bin/ocr.service.worker /go/src/ocr.service.worker
CMD ["/go/src/ocr.service.worker"]