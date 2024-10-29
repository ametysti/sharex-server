FROM golang:1.22.5-alpine AS builder

WORKDIR /go/src/sharex

COPY go.* ./
RUN go mod download

COPY . .

ENV GOCACHE=/root/.cache/go-build

RUN --mount=type=cache,target="/root/.cache/go-build" CGO_ENABLED=0 GOOS=linux go build -v -o /sharex


FROM alpine:latest

WORKDIR /app

COPY --from=builder /sharex /app/sharex

RUN mkdir -p /app/files

EXPOSE 3000

CMD ["./sharex"]