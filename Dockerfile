FROM golang:latest

# Install sudo and wakeonlan
RUN apt-get update && apt-get install -y \
    sudo \
    wakeonlan \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o main.app .

EXPOSE 8080

CMD ["./main.app"]