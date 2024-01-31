FROM golang:1.21.4-alpine3.18 as builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY .. .
RUN go build -o chunker.bin ./cmd

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY .chunks/ .chunks/
COPY ../hack hack/
COPY ../.chunks.yaml .chunks.yaml
COPY --from=builder /app/chunker.bin chunker
ENTRYPOINT ["/app/chunker"]
