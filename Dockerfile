FROM golang:1.9.2-alpine3.7 as builder

RUN apk update && apk upgrade && \
    apk add --no-cache git glide

WORKDIR /go/src/github.com/lordofthejars/testhub

COPY . .

RUN wget https://github.com/gobuffalo/packr/releases/download/v1.11.0/packr_1.11.0_linux_amd64.tar.gz
RUN tar -zxvf packr_1.11.0_linux_amd64.tar.gz 
RUN cp packr /usr/local/bin

RUN glide install
RUN GOOS=linux GOARCH=amd64 packr build -o binaries/testhub

FROM alpine:3.7

RUN addgroup -S testhub && adduser -S -G testhub testhub 
USER testhub

EXPOSE 8000

VOLUME [ "/home/testhub/.hub" ]

WORKDIR /home/testhub
COPY --from=builder /go/src/github.com/lordofthejars/testhub/binaries/testhub .

ENTRYPOINT ["./testhub"]
CMD ["start"]