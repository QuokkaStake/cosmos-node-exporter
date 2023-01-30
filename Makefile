build:
	go build cmd/cosmos-node-exporter.go

install:
	go install cmd/cosmos-node-exporter.go

lint:
	golangci-lint run --fix ./...