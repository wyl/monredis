FROM golang:1.15

RUN mkdir /app

WORKDIR /app

COPY . .

RUN go env -w GOPROXY=https://goproxy.cn,direct

RUN go mod download
