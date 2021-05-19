####################################################################################################
# Step 1: Build the app
####################################################################################################

FROM rwynn/monstache-builder-cache-rel6:1.0.6 AS build-app

RUN mkdir /app

WORKDIR /app

RUN go env -w GOPROXY=https://goproxy.cn,direct

COPY . .

RUN go mod download

RUN make release

####################################################################################################
# Step 2: Copy output build file to an alpine image
####################################################################################################

FROM alpine:3.9.3

ENTRYPOINT ["/bin/monredis"]

EXPOSE 8080

COPY --from=build-app /app/build/linux-amd64/monredis /bin/monredis

