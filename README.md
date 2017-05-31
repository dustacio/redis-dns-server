# Redis Powered DNS Server in GoLang

This is a DNS server that uses Redis as the backend. Redis records are stored
according to the FQDN (with trailing dot) as the key, and a JSON payload as
the value.

## JSON Payload:

```json
{
    "id": 27469,
    "cname": "foo-12345.example.com.",
    "fqdn": "foo-12345.example.com.",
    "ipv4_public_ip": "104.0.0.1",
    "ipv4_private_ip": "",
    "ipv4_private_ip": "10.10.10.1",
    "valid_until": "2015-12-12T03:53:26.150Z"
}
```

Wildcard records, eg. `www.foo-12345.example.com` are supported.  The Redis
key for wildcards is `*.foo-12345.example.com`.

## Usage:

```
./redis-dns-server \
    --domain example.com \
    --redis-server-url redis://127.0.0.1:6379 \
    --port 5300
```

Port `53` is the standard port.  Using a port less than `1024` will require
root privileges.


## Development

### Building

General build steps:

```
$ export GOPATH=/go/src/

$ go get github.com/miekg/dns

$ go get github.com/elcuervo/redisurl

$ go get github.com/hoisie/redis

$ go build -o redis-dns-server redis_dns_server.go main.go

$ ./redis-dns-server --help
```

### Cross Compiling

Reference:

 * http://stackoverflow.com/questions/12168873/cross-compile-go-on-osx

```
cd /usr/local/go/src
sudo GOOS=linux GOARCH=386 CGO_ENABLED=0 ./make.bash --no-clean
```

```
GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -o redis-dns-server.linux
```

### Using Vagrant

A simple `Vagrantfile` is provided to quickly spin up an Ubuntu 14.04 LTS box,
that already has GoLang installed, as well as the latest version of Docker,
and Docker Compose.

```
$ vagrant up

$ vagrant ssh
```

Inside Vagrant, the current working project directory will be accessible at
`/vagrant`.


### Using Docker and Docker Compose

Either from your local machine running Docker, or from within the Vagrant box:

```
$ cp -a docker-compose.env.example docker-compose.env

$ docker-compose up
```

To rebuild the image manually:

```
$ docker-compose build
```


## Deployment

### Docker

Reference the above Development section for using Docker Compose.
Alternatively, Redis DNS Server can be deployed with Docker in the following
fashion.

#### Building

```
$ make linux
$ docker build -t 'redis-dns-server:latest' .
```

#### Linking With a Redis Container

```
$ docker run -tid --name redis redis

$ docker run -itd \
    -e DOMAIN="example.com" \
    -e HOSTNAME="myhostname.example.com" \
    --link redis:db \
    -p 53:53 \
    redis-dns-server:latest
```

#### Linking With an External Redis Server

```
$ docker run -itd \
    -e DOMAIN="example.com" \
    -e HOSTNAME="myhostname.example.com" \
    -e REDIS_HOST="redis.example.com" \
    -p 53:53 \
    redis-dns-server:latest
```

#### Using an Environment File

You can load all `ENV` variables from an `ENV` file.  An example `ENV` file
can be found at `docker-compose.env.example`, and looks something like:

```
REDIS_HOST=redis.example.com
REDIS_PORT=6379
REDIS_DB=0
REDIS_USERNAME=myuser
REDIS_PASSWORD=mypassword
DOMAIN=example.com
DOMAIN_EMAIL=admin.example.com
HOSTNAME=myhostname.example.com
```

Using it:

```
$ docker run -itd \
    --env-file /path/to/myenv.file \
    -p 53:53 \
    redis-dns-server:latest
```

## Inspiration:

 * https://github.com/ConradIrwin/aws-name-server
 * https://github.com/miekg/dns

## TODO List:

 * Use valid_until to calculate the TTL
