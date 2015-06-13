FROM golang:latest
ADD . /go/src/github.com/dustacio/redis-dns-server
RUN go get github.com/miekg/dns
RUN go get github.com/elcuervo/redisurl
RUN go get github.com/hoisie/redis
RUN go install github.com/dustacio/redis-dns-server
ENTRYPOINT ["/go/bin/redis-dns-server", \
            "-domain=${DOMAIN}", \
            "-hostname=${HOSTNAME}", \
            "-redis-server-url=${REDIS_SERVER}"]
EXPOSE 53