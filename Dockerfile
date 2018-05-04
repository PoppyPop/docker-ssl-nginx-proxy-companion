FROM golang:1.9-alpine as builder

RUN set -xe \
	&& apk add --update --no-cache git \
	&& rm -rf /var/cache/apk/*
	
RUN go get github.com/PoppyPop/docker-ssl-nginx-proxy-companion && go install github.com/PoppyPop/docker-ssl-nginx-proxy-companion   

FROM jwilder/docker-gen  
RUN apk add --update --no-cache jq bash curl ca-certificates && \
	 mkdir -p /usr/local/share/ca-certificates/ && \
	 curl -X GET http://yugo.moot.fr:8887/int.pem > /usr/local/share/ca-certificates/int.pem && \
	 curl -X GET http://yugo.moot.fr:8887/ca.pem > /usr/local/share/ca-certificates/ca.pem && \
     update-ca-certificates && \
	 rm /var/cache/apk/* 
	 
COPY --from=builder /go/bin/docker-ssl-nginx-proxy-companion /
ADD notify.sh /notify.sh
