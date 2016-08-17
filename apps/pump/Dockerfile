FROM golang:1.7

RUN go get github.com/Sirupsen/logrus \
    github.com/nats-io/nats \
    gopkg.in/telegram-bot-api.v4


ADD ./src /app
RUN cd /app && make build

EXPOSE 80

WORKDIR /app
ENTRYPOINT ["./pump"]