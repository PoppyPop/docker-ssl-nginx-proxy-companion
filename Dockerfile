FROM golang:1.9-alpine as builder

RUN go get github.com/PoppyPop/docker-ssl-nginx-proxy-companion   


FROM jwilder/docker-gen  

COPY --from=builder /go/src/github.com/PoppyPop/docker/webui-aria2/go-automate-ended/go-automate-ended .
