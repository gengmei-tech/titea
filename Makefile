#
# Makefile
# dev, 2018-04-17 15:05
#

all: build benchmark

build:
	go build -o bin/gm-kv cmd/server/*

benchmark:
	go build -o bin/benchmark cmd/benchmark/main.go


# vim:ft=make
#
