FROM golang:1.9-alpine as builder

RUN go install github.com/PoppyPop/docker-ssl-nginx-proxy-companion   

FROM jwilder/docker-gen  

ADD --from=builder /go/bin/docker-ssl-nginx-proxy-companion /
