FROM golang:1.18-alpine as builder
ENV CGO_ENABLED=0
ENV GOPROXY=https://goproxy.cn/,direct
WORKDIR /go/src/
COPY go.mod go.sum ./
RUN apk add git
RUN go mod download
COPY . .
RUN go build -ldflags '-w -s' -v -o /usr/local/bin/function ./

FROM alpine:latest
COPY --from=builder /usr/local/bin/function /usr/local/bin/function
ENTRYPOINT ["function"]
