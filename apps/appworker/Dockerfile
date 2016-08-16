FROM golang:1.7

RUN go get github.com/nats-io/go-nats-streaming \
    github.com/Sirupsen/logrus \
    github.com/nats-io/nats

ADD ./src /app
RUN cd /app && make build

WORKDIR /app
ENTRYPOINT ["./appworker"]