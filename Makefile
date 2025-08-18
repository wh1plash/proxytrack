build:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/app ./cmd/main.go
	
run: build
	@./bin/app

build-docker:
	@docker build -t proxytrack:latest .

test: 
	@go test -v ./...
	
.PHONY: build