FROM golang:1.22-alpine as go-builder


WORKDIR /app
ADD /code /app

RUN GOPROXY=https://goproxy.cn go mod download

RUN CGO_ENABLED=0 go build -ldflags '-w -s' -a -o  feishu-openai

FROM alpine:3.20
RUN apk --no-cache add tzdata && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && echo "Asia/Shanghai" >/etc/timezone

WORKDIR /app

# RUN apk add --no-cache bash
COPY --from=go-builder /app/feishu-openai /app
EXPOSE 9000
ENTRYPOINT ["/app/feishu-openai"]
