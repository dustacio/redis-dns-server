package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/elcuervo/redisurl"
	"github.com/hoisie/redis"
)

func main() {
	flag.Usage = func() {
		fmt.Printf("\nUsage: redis-dns-server -domain <domain> -redis-server-url <redis-server-url>\n\n")
		flag.PrintDefaults()
		fmt.Printf("\nThe dns-server needs permission to bind to given port, default is 53.\n\n")
	}

	domain := flag.String("domain", "", "Domain for which the server serves")
	redisServerURLStr := flag.String("redis-server-url", "", "redis://[:password]@]host:port[/db-number][?option=value]")
	hostname := flag.String("hostname", "", "Public hostname of *this* server")
	port := flag.Int("port", 53, "Port")
	help := flag.Bool("help", false, "Get help")
	emailAddr := "admin." + *domain
	mbox := flag.String("mbox", emailAddr, "Domain Admin Email Address")

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	} else if *domain == "" || *redisServerURLStr == "" {
		flag.Usage()
		fmt.Println("  -domain and -redis-server-url are required parameters")
		os.Exit(1)
	}

	if strings.Contains(*mbox, "@") {
		fmt.Println("Email addresses in DNS can not contain the character @")
		os.Exit(0)
	}

	if *hostname == "" {
		*hostname, _ = os.Hostname()
	}

	redisClient := RedisClient(*redisServerURLStr)
	server := NewRedisDNSServer(*domain, *hostname, redisClient, *mbox)
	portStr := fmt.Sprintf(":%d", *port)
	log.Printf("Serving DNS records for *.%s from %s port %s", server.domain,
		server.hostname, portStr)

	go checkNSRecordMatches(server.domain, server.hostname)

	go server.listenAndServe(portStr, "udp")
	server.listenAndServe(portStr, "tcp")
}

// RedisClient is a client to the Redis server given by urlStr
func RedisClient(urlStr string) redis.Client {
	url := redisurl.Parse(urlStr)

	fmt.Println("HOST: ", url.Host)
	fmt.Println("DB: ", url.Database)

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
