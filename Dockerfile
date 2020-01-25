FROM golang:1.13-alpine AS builder
RUN apk add libcec-dev eudev-dev p8-platform-dev git build-base
WORKDIR /cec
COPY go.mod go.sum ./
RUN go mod download
COPY . /cec
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o cec *.go
FROM alpine
RUN apk add libcec eudev p8-platform
COPY --from=builder /cec/cec /cec
ENTRYPOINT ["/cec"]