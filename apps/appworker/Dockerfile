FROM golang:1.7

RUN go get github.com/Sirupsen/logrus \
    github.com/nats-io/nats \
    gopkg.in/telegram-bot-api.v4 \
    gopkg.in/pg.v4 \
    github.com/inpime/sdata

ADD ./src /app/src
RUN cd /app/src && make build

WORKDIR /app/src
ENTRYPOINT ["./appworker"]