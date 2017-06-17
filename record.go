package main

import "net"

// Record is the json format message that is stored in Redis
type Record struct {
	CNames        []string `json:"cnames"`
	IPv4PublicIPs []net.IP `json:"ipv4_public_ips"`
	IPv6PublicIPs []net.IP `json:"ipv6_public_ips"`
	MBox          string   `json:"mbox"`
	MXServers     []string `json:"mx_servers"`
	NameServers   []string `json:"name_servers"`
	SOA           string   `json:"soa"`
	TTL           uint32   `json:"ttl"`
}
