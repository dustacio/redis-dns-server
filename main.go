package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/elcuervo/redisurl"
	"github.com/hoisie/redis"
)

func main() {
	flag.Usage = func() {
		fmt.Printf("\nUsage: redis-dns-server -redis-server-url <redis-server-url> -mbox <admin.mail.box.no.at.sign>\n\n")
		flag.PrintDefaults()
		fmt.Printf("\nThe dns-server needs permission to bind to given port, default is 53.\n\n")
	}

	redisServerURLStr := flag.String("redis-server-url", "", "redis://[:password]@]host:port[/db-number][?option=value]")
	domain := flag.String("domain", "", "No longer used")
	hostname := flag.String("hostname", "", "Public hostname of *this* server")
	port := flag.Int("port", 53, "Port")
	help := flag.Bool("help", false, "Get help")
	mbox := flag.String("mbox", "", "Domain Admin Email Address")

	flag.Parse()

	if *domain != "" {
		fmt.Println("Domain provided but no longer used", *domain)
	}

	if *help {
		flag.Usage()
		os.Exit(0)
	} else if *mbox == "" || *redisServerURLStr == "" {
		flag.Usage()
		fmt.Println("  -mbox and -redis-server-url are required parameters")
		os.Exit(1)
	}

	if strings.Contains(*mbox, "@") {
		fmt.Println("Email addresses in DNS can not contain the @ character, use dots")
		os.Exit(0)
	}

	if *hostname == "" {
		*hostname, _ = os.Hostname()
	}

	log.Printf("Redis: %s\n", *redisServerURLStr)
	redisClient := RedisClient(*redisServerURLStr)
	server := NewRedisDNSServer(*hostname, redisClient, *mbox)
	portStr := fmt.Sprintf("0.0.0.0:%d", *port)
	log.Printf("Serving DNS records from %s port %s",
		server.hostname, portStr)

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
