#!/bin/sh

# ARE WE LINKED TO A DOCKER CONTAINER?
if [ ! -z $DB_PORT_6379_TCP_ADDR ]; then
    REDIS_HOST=$DB_PORT_6379_TCP_ADDR
    REDIS_PORT=$DB_PORT_6379_TCP_PORT
fi

[ -z $REDIS_HOST ] && echo "REDIS_HOST Not Set" && exit 1
[ -z $REDIS_PORT ] && REDIS_PORT="6379"
[ -z $REDIS_DB ] && REDIS_DB="0"
[ -z $DOMAIN ] && echo "DOMAIN Not Set" && exit 1
[ -z $DOMAIN_EMAIL ] && DOMAIN_EMAIL="admin@${DOMAIN}"
[ -z $HOSTNAME ] && echo "HOSTNAME Not Set" && exit 1

if [ ! -z $REDIS_USERNAME ]; then
    USER_PASS="${REDIS_USERNAME}:${REDIS_PASSWORD}@"
else
    USER_PASS=""
fi

URI="redis://${USER_PASS}${REDIS_HOST}:${REDIS_PORT}/${REDIS_DB}"
/go/bin/redis-dns-server \
    --domain=${DOMAIN} \
    --mbox=${DOMAIN_EMAIL} \
    --hostname=${HOSTNAME} \
    --redis-server-url=${URI}