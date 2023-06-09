# syntax = docker/dockerfile:1.2
FROM golang:1.18
WORKDIR /app

# for cacheing purpose
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build go build -o main
EXPOSE 8080
CMD ["./main"]
