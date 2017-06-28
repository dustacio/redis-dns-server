package main

import (
	"log"
	"strconv"

	"github.com/miekg/dns"
)

// Answer crafts a response to the DNS Question
func (s *RedisDNSServer) Answer(msg dns.Question) []dns.RR {
	var answers []dns.RR
	record := s.Get(msg.Name)

	// Bail out early if the record isn't found in the key store
	if record == nil {
		log.Printf("Error no record found for %s\n", msg.Name)
		return nil
	}

	switch msg.Qtype {
	case dns.TypeNS:
		answers = append(answers, NS(msg.Name, record)...)
	case dns.TypeSOA:
		answers = append(answers, SOA(msg.Name, record, s.getSerialNumber()))
	// case dns.TypeA:
	// 	answers = append(answers, A(msg.Name, record)...)
	// case dns.TypeAAAA:
	// 	answers = append(answers, AAAA(msg.Name, record)...)
	// case dns.TypeCNAME:
	// 	answers = append(answers, CNAME(msg.Name, record)...)
	case dns.TypeMX:
		answers = append(answers, MX(msg.Name, record)...)
	default:
		answers = append(answers, s.Host(msg.Name, record)...)
	}
	return answers
}

func (s *RedisDNSServer) getSerialNumber() uint32 {
	sn, err := s.redisClient.Get(serialNumberKey)
	if err != nil {
		log.Printf("Error reading SerialNumber: %s\n", err)
		return uint32(0)
	}

	x, _ := strconv.Atoi(string(sn))
	log.Printf("Found SerialNumber %d\n", uint32(x))
	return uint32(x)
}

// SOA returns the Server of Authority record response
func SOA(name string, record *Record, serialNumber uint32) dns.RR {
	return &dns.SOA{
		Hdr:     dns.RR_Header{Name: name, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: 60},
		Ns:      dns.Fqdn(record.NameServers[0]),
		Mbox:    record.MBox,
		Serial:  serialNumber,
		Refresh: 86400,
		Retry:   7200,
		Expire:  3600, // RFC1912 suggests 2-4 weeks 1209600-2419200
		Minttl:  60,
	}
}

// NS returns the Name Servers
func NS(name string, record *Record) []dns.RR {
	var answers []dns.RR
	for i := 0; i < len(record.NameServers); i++ {
		r := new(dns.NS)
		r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeNS, Class: dns.ClassINET, Ttl: record.TTL}
		r.Ns = dns.Fqdn(record.NameServers[i])
		answers = append(answers, r)
	}
	return answers
}

// A returns A records
func A(name string, record *Record) []dns.RR {
	var answers []dns.RR
	for i := 0; i < len(record.IPv4PublicIPs); i++ {
		r := new(dns.A)
		r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: record.TTL}
		r.A = record.IPv4PublicIPs[i]
		answers = append(answers, r)
	}
	return answers
}

// AAAA returns IPv6 records
func AAAA(name string, record *Record) []dns.RR {
	var answers []dns.RR
	for i := 0; i < len(record.IPv4PublicIPs); i++ {
		r := new(dns.AAAA)
		r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: record.TTL}
		r.AAAA = record.IPv4PublicIPs[i]
		answers = append(answers, r)
	}
	return answers
}

// CNAME returns canonical names
func CNAME(name string, record *Record) []dns.RR {
	var answers []dns.RR
	for i := 0; i < len(record.CNames); i++ {
		r := new(dns.CNAME)
		r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeCNAME, Class: dns.ClassINET, Ttl: record.TTL}
		r.Target = dns.Fqdn(record.CNames[i])
		answers = append(answers, r)
	}
	return answers
}

// MX returns mail records
func MX(name string, record *Record) []dns.RR {
	var answers []dns.RR
	for i := 0; i < len(record.MXServers); i++ {
		r := new(dns.MX)
		r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeMX, Class: dns.ClassINET, Ttl: record.TTL}
		r.Mx = record.MXServers[i]
		answers = append(answers, r)
	}
	return answers

}

// Host returns cname, a, and aaaa in that order
func (s *RedisDNSServer) Host(name string, record *Record) []dns.RR {
	var answers []dns.RR

	answers = append(answers, CNAME(name, record)...)
	answers = append(answers, A(name, record)...)
	answers = append(answers, AAAA(name, record)...)

	return answers
}
