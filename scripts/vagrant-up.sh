#!/bin/bash

wget -qO- https://get.docker.io/gpg | sudo apt-key add -
sh -c "echo deb http://get.docker.io/ubuntu docker main > /etc/apt/sources.list.d/docker.list"
apt-get update && apt-get upgrade -y

# install latest docker, and redis-server (for testing external redis conn)
apt-get install -y redis-server lxc-docker
pip install -U docker-compose

exit 0
