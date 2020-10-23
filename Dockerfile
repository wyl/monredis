####################################################################################################
# Step 1: Build the app
####################################################################################################

FROM golang:1.15 AS build-app

RUN mkdir /app

WORKDIR /app

COPY . .

RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod download

RUN make release

####################################################################################################
# Step 2: Copy output build file to an alpine image
####################################################################################################

FROM alpine:3.9.3

ENTRYPOINT ["/bin/monredis"]

COPY --from=build-app /app/build/linux-amd64/monredis /bin/monredis
