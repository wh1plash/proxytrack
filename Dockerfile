FROM golang:1.24

WORKDIR /app

COPY go.* ./

RUN apt-get update && apt-get upgrade -y && go mod download

COPY . .

RUN make build

EXPOSE 8084

CMD ["./bin/app"]