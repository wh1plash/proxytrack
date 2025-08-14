build:
	@go build -o bin/app ./cmd/main.go
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/app ./cmd/main.go
	
run: build
	@./bin/app

test: 
	@go test -v ./...
	
.PHONY: build