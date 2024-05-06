FROM golang:1.16 AS builder

WORKDIR /app

COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/server /app/server
COPY --from=builder /app/storage.pb.go /app/storage.pb.go

CMD ["/app/server"]