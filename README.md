Redis Powered DNS Server in golang

This is a DNS server that uses Redis as the backend. Redis
records are stored according to the fqdn (with trailing dot)
as the key, and a JSON payload as the value.


JSON Payload:
```

{"id":27469,
 "cname":"as-12345.ascreen.co.",
 "fqdn":"as-12345.ascreen.co.",
 "public_ip":"104.0.0.1",
 "private_ip":"10.10.10.1",
 "valid_until":"2015-12-12T03:53:26.150Z"
 }

```

Wildcard records, eg. www.as-12345.ascreen.co are supported,
Redis is key for wildcards is *.as-12345.ascreen.co

Usage:
```
./redis-dns-server --domain ascreen.co --redis-server-url redis://127.0.0.1:6379 --port 5300

53 is the standard port, ports less than 1024 require root privileges.
```

Inspiration:

https://github.com/ConradIrwin/aws-name-server
https://github.com/miekg/dns

TODO:
Use valid_until to calculate the TTL

