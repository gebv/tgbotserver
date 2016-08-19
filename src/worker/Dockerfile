FROM golang:1.7

RUN go get -v github.com/Sirupsen/logrus \
    github.com/nats-io/nats \
    gopkg.in/telegram-bot-api.v4 \
    gopkg.in/pg.v4 \
    github.com/inpime/sdata \
    github.com/satori/go.uuid

ADD ./src /usr/local/src/app/src
RUN cd /usr/local/src/app/src && \
    GOPATH=${GOPATH}:${PWD}/.. \
    CGO_ENABLED=0 \
    go build -o /usr/local/bin/app \
    -v -a --installsuffix cgo -ldflags "-s" \
    main.go db.go app.go

WORKDIR /usr/local/src/app
ENTRYPOINT ["app"]