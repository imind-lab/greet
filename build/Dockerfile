FROM golang:alpine as builder
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
RUN apk --no-cache add build-base gcc git openssh binutils-gold
WORKDIR /go/src/github.com/imind-lab/greeter/
COPY . .
RUN go get ./...
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -installsuffix cgo -o greeter main.go
RUN go get github.com/grpc-ecosystem/grpc-health-probe && go install github.com/grpc-ecosystem/grpc-health-probe

FROM alpine:latest
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
RUN apk add --no-cache tzdata \
    && ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && rm -rf /var/cache/apk/* /tmp/* /var/tmp/* $HOME/.cache

WORKDIR .
ADD conf /conf
COPY --from=builder /go/src/github.com/imind-lab/greeter /go/bin/grpc-health-probe /bin/
ENTRYPOINT [ "/bin/greeter" ]
