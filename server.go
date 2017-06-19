package main

import (
	"log"
	"strings"

	"github.com/hoisie/redis"
	"github.com/miekg/dns"
)

// TTL Time to Live in seconds, default value
const TTL uint32 = 300

// The key used to read the serial number
const serialNumberKey = "redis-dns-server-serial-no"

// RedisDNSServer contains the configuration details for the server
type RedisDNSServer struct {
	hostname    string
	redisClient redis.Client
	mbox        string
}

// NewRedisDNSServer is a convienence for creating a new server
func NewRedisDNSServer(hostname string, redisClient redis.Client, mbox string) *RedisDNSServer {
	if !strings.HasSuffix(hostname, ".") {
		hostname += "."
	}

	server := &RedisDNSServer{
		hostname:    hostname,
		redisClient: redisClient,
	}

	dns.HandleFunc(".", server.handleRequest)
	return server
}

func (s *RedisDNSServer) listenAndServe(port, net string) {
	server := &dns.Server{Addr: port, Net: net}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("%s", err)
	}
}
