FROM golang:latest
ADD . /go/src/github.com/dustacio/redis-dns-server
RUN go get github.com/miekg/dns
RUN go get github.com/elcuervo/redisurl
RUN go get github.com/hoisie/redis
RUN go install github.com/dustacio/redis-dns-server

# ENV is not parsed in CMD/ENTRYPOINT?
ENTRYPOINT ["/go/src/github.com/dustacio/redis-dns-server/scripts/docker-entrypoint.sh"]
EXPOSE 53