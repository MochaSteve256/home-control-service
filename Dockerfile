FROM golang:latest

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o main.app .

EXPOSE 8080

CMD ["./main.app"]