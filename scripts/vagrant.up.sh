#!/bin/bash

apt-get update && apt-get upgrade -y
apt-get install -y docker.io redis-server 
service redis-server restart

exit 0
