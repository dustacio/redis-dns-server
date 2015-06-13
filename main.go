package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/elcuervo/redisurl"
	"github.com/hoisie/redis"
)

const USAGE = `Usage: redis-dns-server --domain <domain>
                    --redis-server-url <redis-server-url>
                    [ 
                    --port <port>
                    --hostname <hostname>
                    --mbox <domainemailaddress>
                    ]

The dns-server needs permission to bind to given port, default is 53.

`

func main() {
	domain := flag.String("domain", "", "Domain for which the server serves")
	redisServerURLStr := flag.String("redis-server-url", "", "redis://[:password]@]host:port[/db-number][?option=value]")
	hostname := flag.String("hostname", "", "Public hostname of *this* server")
	port := flag.Int("port", 53, "Port")
	help := flag.Bool("help", false, "Get help")
	emailAddr := "admin@" + *domain
	mbox := flag.String("mbox", emailAddr, "Domain Admin Email Address")

	flag.Parse()

	if *domain == "" {
		fmt.Println(USAGE)
		log.Fatalf("missing required parameter: --domain")
	} else if *redisServerURLStr == "" {
		fmt.Println(USAGE)
		log.Fatalf("missing required parameter: --redis-server-url")
	} else if *help {
		fmt.Println(USAGE)
		os.Exit(0)
	}

	if *hostname == "" {
		*hostname, _ = os.Hostname()
	}

	redisClient := RedisClient(*redisServerURLStr)
	server := NewRedisDNSServer(*domain, *hostname, redisClient, *mbox)
	port_str := fmt.Sprintf(":%d", *port)
	log.Printf("Serving DNS records for *.%s from %s port %s", server.domain,
		server.hostname, port_str)

	go checkNSRecordMatches(server.domain, server.hostname)

	go server.listenAndServe(port_str, "udp")
	server.listenAndServe(port_str, "tcp")
}

func RedisClient(urlStr string) redis.Client {
	url := redisurl.Parse(urlStr)
	var client redis.Client
	address := fmt.Sprintf("%s:%d", url.Host, url.Port)
	client.Addr = address
	client.Db = url.Database
	log.Printf("Redis DB is %d", url.Database)
	client.Password = url.Password

	return client
}

func checkNSRecordMatches(domain, hostname string) {
	time.Sleep(1 * time.Second)
	results, err := net.LookupNS(domain)

	if err != nil {
		log.Printf("No working NS records found for %s", domain)
	}

	matched := false
	if len(results) > 0 {
		for _, record := range results {
			if record.Host == hostname {
				matched = true
			}
		}
		if !matched {
			log.Printf("The NS record for %s is %s", domain, results[0].Host)
			log.Printf(" --hostname is %s", hostname)
			log.Printf("These must match for DNS to work")
		}
	}
}
