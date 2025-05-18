FROM golang:1.24
RUN apt-get update && apt-get install -y --no-install-recommends socat

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /usr/local/bin/server -v ./cmd/server

ENTRYPOINT ["sh", "-c", "socat TCP-LISTEN:1025,reuseaddr,fork TCP:mailhog:1025 & exec server"]
