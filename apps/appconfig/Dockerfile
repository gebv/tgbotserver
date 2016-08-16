FROM golang:1.7

RUN go get github.com/Sirupsen/logrus \
    gopkg.in/telegram-bot-api.v4

ADD ./src /app
RUN cd /app && make build

WORKDIR /app
ENTRYPOINT ["./appconfig"]