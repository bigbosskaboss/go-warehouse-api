FROM golang:latest

WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download

COPY . .

ENV CONFIG_PATH="./config/docker.yaml"

RUN go build -o main ./cmd/warehouse/main.go

CMD ["/app/main"]
