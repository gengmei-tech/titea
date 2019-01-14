# Builder image
FROM golang:1.10.1-alpine as builder

RUN apk add --no-cache \
    make \
    git

COPY . /go/src/github.com/gengmei-tech/titea

WORKDIR /go/src/github.com/gengmei-tech/titea

RUN env GOOS=linux CGO_ENABLED=0 make

# Executable image
FROM scratch

COPY --from=builder /go/src/github.com/gengmei-tech/titea/bin/titea /titea/bin/titea
COPY --from=builder /go/src/github.com/gengmei-tech/titea/config/titea.config.toml /titea/conf/titea.config.toml

WORKDIR /titea

EXPOSE 5379

ENTRYPOINT ["./bin/titea"]