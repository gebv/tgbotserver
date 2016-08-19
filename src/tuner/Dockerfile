FROM golang:1.7

RUN go get -v github.com/Sirupsen/logrus \
    gopkg.in/telegram-bot-api.v4

ADD ./ /usr/local/src/app/src
RUN cd /usr/local/src/app/src && \
    GOPATH=${GOPATH}:${PWD}/.. \
    CGO_ENABLED=0 \
    go build -o /usr/local/bin/app \
    -v -a --installsuffix cgo -ldflags "-s" \
    main.go

WORKDIR /usr/local/src/app/src
ENTRYPOINT ["app"]