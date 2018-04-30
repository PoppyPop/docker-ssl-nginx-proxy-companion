FROM golang:1.9-alpine as builder

RUN set -xe \
	&& apk update --no-cache && apk upgrade --no-cache \
	&& apk add --update --no-cache git \
	&& rm -rf /var/cache/apk/*
	
RUN go get github.com/PoppyPop/docker-ssl-nginx-proxy-companion && go install github.com/PoppyPop/docker-ssl-nginx-proxy-companion   


FROM jwilder/docker-gen  
COPY --from=builder /go/bin/docker-ssl-nginx-proxy-companion /
