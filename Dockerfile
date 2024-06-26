FROM golang:1.21-alpine AS builder

WORKDIR /src

COPY . .

RUN go build -o /go/bin/fts .

FROM alpine:latest

COPY --from=builder /go/bin/fts /bin/fts