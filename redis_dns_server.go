package main

import (
	"encoding/json"
	"github.com/hoisie/redis"
	"github.com/miekg/dns"
	"log"
	"net"
	"strings"
	"time"
)

// TTL Time to Live in seconds
const TTL uint32 = 300

// Record is the json format message that is stored in Redis
type Record struct {
	CName         string    `json:"cname"`
	PublicIP      net.IP    `json:"public_ip"`
	PrivateIP     net.IP    `json:"private_ip"`
	ValidUntil    time.Time `json:"valid_until"`
	IPv4PublicIP  net.IP    `json:"ipv4_public_ip"`
	IPv6PublicIP  net.IP    `json:"ipv6_public_ip"`
	IPv4PrivateIP net.IP    `json:"ipv4_private_ip"`
}

// RedisDNSServer contains the configuration details for the server
type RedisDNSServer struct {
	domain      string
	hostname    string
	redisClient redis.Client
	mbox        string
}

// response is the dns message
type response struct {
	*dns.Msg
}

// NewRedisDNSServer is a convienence for creating a new server
func NewRedisDNSServer(domain string, hostname string, redisClient redis.Client, mbox string) *RedisDNSServer {
	if !strings.HasSuffix(domain, ".") {
		domain += "."
	}

	if !strings.HasSuffix(mbox, ".") {
		mbox += "."
	}

	if !strings.HasSuffix(hostname, ".") {
		hostname += "."
	}

	server := &RedisDNSServer{
		domain:      domain,
		hostname:    hostname,
		redisClient: redisClient,
		mbox:        mbox,
	}

	dns.HandleFunc(server.domain, server.handleRequest)
	return server
}

func (s *RedisDNSServer) listenAndServe(port, net string) {
	server := &dns.Server{Addr: port, Net: net}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("%s", err)
	}
}

func (s *RedisDNSServer) handleRequest(w dns.ResponseWriter, request *dns.Msg) {
	r := new(dns.Msg)
	r.SetReply(request)
	r.Authoritative = true

	for _, msg := range request.Question {
		log.Printf("%v %#v %v (id=%v)", dns.TypeToString[msg.Qtype], msg.Name, w.RemoteAddr(), request.Id)

		answers := s.Answer(msg)

		if len(answers) > 0 {
			log.Println("Answers present")
			r.Answer = append(r.Answer, answers...)
		} else {
			log.Println("No Answers present")
			r.Answer = append(r.Ns, s.SOA(msg))
			log.Println("Appended...")
		}
	}
	log.Printf("Send Reply: %+v", r)
	w.WriteMsg(r)
	log.Printf("SentReply")
}

// Answer crafts a response to the DNS Question
func (s *RedisDNSServer) Answer(msg dns.Question) []dns.RR {
	var answers []dns.RR
	record := s.Lookup(msg)
	ttl := TTL
	// Bail out early if the record isn't found in the key store
	if record == nil || record.CName == "" {
		log.Println("No record in key store, returning nil")
		return nil
	}

	switch msg.Qtype {
	case dns.TypeNS:
		log.Println("Processing NS request")
		if msg.Name == s.domain {
			answers = append(answers, &dns.NS{
				Hdr: dns.RR_Header{Name: msg.Name, Rrtype: dns.TypeNS, Class: dns.ClassINET, Ttl: 300},
				Ns:  s.hostname,
			})
		}
	case dns.TypeSOA:
		log.Println("Processing SOA request")
		log.Println("  Domain is", s.domain)
		if msg.Name == s.domain {
			log.Println("  msg.Name == s.domain", msg.Name, s.domain)

			r := new(dns.SOA)
			r.Hdr = dns.RR_Header{Name: msg.Name, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: 60}
			r.Ns = s.hostname
			r.Mbox = s.mbox
			r.Serial = uint32(time.Now().Unix())
			r.Refresh = 60
			r.Retry = 60
			r.Expire = 86400
			r.Minttl = 60
			answers = append(answers, r)
		} else {
			log.Println("  no match!", msg.Name, s.domain)
		}
	case dns.TypeA:
		log.Println("Processing A request")
		addr := record.IPv4PublicIP
		// bail if no ip address
		if addr == nil {
			log.Println("No ip address, returning nil")
			return nil
		}
		r := new(dns.A)
		r.Hdr = dns.RR_Header{Name: msg.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: ttl}
		r.A = addr
		answers = append(answers, r)
	case dns.TypeAAAA:
		log.Println("Processing AAAA request")
		addr := record.IPv6PublicIP
		// bail if no ip address
		if addr == nil {
			return nil
		}

		r := new(dns.AAAA)
		r.Hdr = dns.RR_Header{Name: msg.Name, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: ttl}
		r.AAAA = addr
		answers = append(answers, r)
	case dns.TypeCNAME:
		log.Println("Processing CNAME request")
		r := new(dns.CNAME)
		r.Hdr = dns.RR_Header{Name: msg.Name, Rrtype: dns.TypeCNAME, Class: dns.ClassINET, Ttl: ttl}
		r.Target = msg.Name
		answers = append(answers, r)
	case dns.TypeMX:
		log.Println("Processing MX request")
		r := new(dns.MX)
		r.Hdr = dns.RR_Header{Name: msg.Name, Rrtype: dns.TypeMX, Class: dns.ClassINET, Ttl: ttl}
		r.Preference = 10
		r.Mx = msg.Name
		answers = append(answers, r)
	}
	return answers
}

// Lookup will locate the details in Redis for the fqdn, if not found
// lookup will try to locate a wildcard entry for the fqdn
func (s *RedisDNSServer) Lookup(msg dns.Question) *Record {
	log.Printf("LOOKUP: Looking for '%s'\n", msg.Name)
	str, err := s.redisClient.Get(msg.Name)
	log.Printf(" found str\n%s", str)
	var result Record

	// error indicates that the record was not found
	if err != nil {
		wildcard := wildCardHostName(msg.Name)
		log.Printf("No record for %s, trying wildcard %s\n", msg.Name, wildcard)

		domainDots := strings.Count(s.domain, ".") + 1
		msgDots := strings.Count(msg.Name, ".")
		if msgDots <= domainDots {
			log.Printf("msgDots <= domainDots returning nil")
			return nil
		}

		str, err = s.redisClient.Get(wildcard)
		if err != nil {
			log.Printf("No record for %s\n", wildcard)
			return nil
		}
	}
	json.Unmarshal([]byte(str), &result)
	return &result
}

func wildCardHostName(hostName string) string {
	nameParts := strings.SplitAfterN(hostName, ".", 2)
	return "*." + nameParts[1]
}

// SOA returns the Server of Authority record response
func (s *RedisDNSServer) SOA(msg dns.Question) dns.RR {
	r := new(dns.SOA)
	r.Hdr = dns.RR_Header{Name: s.domain, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: 60}
	r.Ns = s.hostname
	r.Mbox = s.mbox
	r.Serial = uint32(time.Now().Unix())
	r.Refresh = 86400
	r.Retry = 7200
	r.Expire = 86400
	r.Minttl = 60
	return r
}
