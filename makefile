.PHONY: all
TARGET := lottery_singlesvr
GOENV := GOOS=linux GOARCH=amd64
GOMACENV := GOOS=darwin GOARCH=amd64
all:
	cd cmd/
	CGO_ENABLED=0 ${GOENV} go build -o ../bin/${TARGET}

clean:
	rm -rf ${TARGET}
format:
	gofmt -w .
	goimports -w .
	golint ./...
test:
	go test --cover -gcflags=-l ./...

build:
	@./scripts/build.sh

start:
	@./scripts/server.sh start

stop:
	@./scripts/server.sh stop

restart:
	@./scripts/server.sh restart