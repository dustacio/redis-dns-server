FROM alpine
ADD redis-dns-server.linux /redis-dns-server
ADD scripts/docker-entrypoint.sh /docker-entrypoint.sh

# ENV is not parsed in CMD/ENTRYPOINT?
# ENTRYPOINT ["/go/src/github.com/dustacio/redis-dns-server/scripts/docker-entrypoint.sh"]
ENTRYPOINT ["/docker-entrypoint.sh"]
EXPOSE 53