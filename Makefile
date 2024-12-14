APPS = step1

# CFLAGS := $(CFLAGS) -g -W -Wall -Wno-unused-parameter -iquote .

.PHONY: all test clean

all: build

build:
	mkdir -p bin
	go build -o ./bin/$(APPS) .

test:
	./bin/$(APPS)

clean:
	rm -rf ./bin/*
