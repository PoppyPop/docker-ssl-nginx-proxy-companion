NAME = poppypop/docker-ssl-nginx-proxy-companion
VERSION = latest

.PHONY: all build run shell

all: build

build:	
	docker build -t $(NAME):$(VERSION) --rm .

run:
	docker run -ti --rm $(NAME):$(VERSION)
	
shell:
	docker run -ti --rm $(NAME):$(VERSION) /bin/sh