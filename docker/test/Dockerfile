FROM rwynn/monstache-builder-cache-rel6:1.0.5

RUN mkdir /app

WORKDIR /app

COPY . .

CMD go env -w GOPROXY=https://goproxy.cn,direct

RUN go mod download
