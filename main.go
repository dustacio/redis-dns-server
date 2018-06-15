package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

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
	mbox := flag.String("mbox", "", "No longer used")

	Header()
	flag.Parse()

	if *domain != "" {
		log.Println("Domain provided but no longer used", *domain)
	}

	if *mbox != "" {
		log.Println("Mbox provided but no longer used", *mbox)
	}

	if *help {
		flag.Usage()
		os.Exit(0)
	} else if *redisServerURLStr == "" {
		flag.Usage()
		log.Println("  -redis-server-url is a required parameter")
		os.Exit(1)
	}

	if *hostname == "" {
		*hostname, _ = os.Hostname()
	}

	redisClient := RedisClient(*redisServerURLStr)
	server := NewRedisDNSServer(*hostname, redisClient, *mbox)
	hostPort := net.JoinHostPort("0.0.0.0", strconv.Itoa(*port))
	log.Printf("Serving DNS records from %s on %s", server.hostname, hostPort)

	go server.listenAndServe(hostPort, "udp")
	server.listenAndServe(hostPort, "tcp")
}

// Header fancy ascii art
func Header() {
	head := `
     ___                         ___           ___
    /\  \         _____         /\  \         /\__\
   /::\  \       /::\  \        \:\  \       /:/ _/_
  /:/\:\__\     /:/\:\  \        \:\  \     /:/ /\  \
 /:/ /:/  /    /:/  \:\__\   _____\:\  \   /:/ /::\  \
/:/_/:/__/___ /:/__/ \:|__| /::::::::\__\ /:/_/:/\:\__\
\:\/:::::/  / \:\  \ /:/  / \:\~~\~~\/__/ \:\/:/ /:/  /
 \::/~~/~~~~   \:\  /:/  /   \:\  \        \::/ /:/  /
  \:\~~\        \:\/:/  /     \:\  \        \/_/:/  /
   \:\__\        \::/  /       \:\__\         /:/  /
    \/__/         \/__/         \/__/         \/__/

     Redis DNS Server Version: Strange Days 1.0.100
`
	log.Println(head)

}

// RedisClient is a client to the Redis server given by urlStr
func RedisClient(urlStr string) redis.Client {
	url := redisurl.Parse(urlStr)

	log.Printf("Redis Client Host: %s, Port: %d, DB: %d\n", url.Host, url.Port, url.Database)

	var client redis.Client
	address := net.JoinHostPort(url.Host, strconv.Itoa(url.Port))
	client.Addr = address
	client.Db = url.Database
	client.Password = url.Password

	return client
}
