test:
	go test -v -bench=. -benchmem -run=. ./...
.PHONY: test

testwithrace:
	go test -v -race -bench=. -benchmem -run=. ./...
.PHONY: testwithraced

build:
	GOPATH=${GOPATH}:${PWD}/.. CGO_ENABLED=0 go build -o appworker -v -a --installsuffix cgo -ldflags "-s" main.go db.go handler.go 
.PHONY: build