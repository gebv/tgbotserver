test:
	go test -v -bench=. -benchmem -run=. ./...
.PHONY: test

testwithrace:
	go test -v -race -bench=. -benchmem -run=. ./...
.PHONY: testwithraced

build:
	CGO_ENABLED=0 go build -o pump -v -a --installsuffix cgo -ldflags "-s" main.go
.PHONY: build