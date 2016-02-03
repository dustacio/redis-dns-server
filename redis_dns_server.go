package main

import (
	"encoding/json"
	"fmt"
	"github.com/hoisie/redis"
	"github.com/miekg/dns"
	"log"
	"net"
	"strings"
	"time"
)

const TTL uint32 = 300

type Record struct {
	CName      string    `json:"cname"`
	PublicIP   net.IP    `json:"public_ip"`
	PrivateIP  net.IP    `json:"private_ip"`
	ValidUntil time.Time `json:"valid_until"`
}

type RedisDNSServer struct {
	domain      string
	hostname    string
	redisClient redis.Client
	mbox        string
}

type response struct {
	*dns.Msg
}

func NewRedisDNSServer(domain string, hostname string, redisClient redis.Client, mbox string) *RedisDNSServer {
	if !strings.HasSuffix(domain, ".") {
		domain += "."
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
			r.Answer = append(r.Answer, answers...)
		} else {
			r.Ns = append(r.Ns, s.SOA(msg))
		}
	}
	w.WriteMsg(r)
}

func (s *RedisDNSServer) Answer(msg dns.Question) (answers []dns.RR) {
	if msg.Qtype == dns.TypeNS {
		if msg.Name == s.domain {
			answers = append(answers, &dns.NS{
				Hdr: dns.RR_Header{Name: msg.Name, Rrtype: dns.TypeNS, Class: dns.ClassINET, Ttl: 300},
				Ns:  s.hostname,
			})
		}
		return answers
	}
	if msg.Qtype == dns.TypeSOA {
		if msg.Name == s.domain {
			answers = append(answers, s.SOA(msg))
		}
		return answers
	}

	record := s.Lookup(msg)
	ttl := TTL

	if msg.Qtype == dns.TypeCNAME {
		fmt.Println("CNAME request")
		answers = append(answers, &dns.CNAME{
			Hdr:    dns.RR_Header{Name: msg.Name, Rrtype: dns.TypeCNAME, Class: dns.ClassINET, Ttl: ttl},
			Target: record.CName,
		})
	} else if msg.Qtype == dns.TypeA {
		fmt.Println("A request")
		addr := record.PublicIP
		if record.PublicIP != nil {
			addr = record.PrivateIP
		}
		answers = append(answers, &dns.A{
			Hdr: dns.RR_Header{Name: msg.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: ttl},
			A:   addr,
		})
	} else if msg.Qtype == dns.TypeMX {
		fmt.Println("MX request")
		answers = append(answers, &dns.MX{
			Hdr:        dns.RR_Header{Name: msg.Name, Rrtype: dns.TypeMX, Class: dns.ClassINET, Ttl: ttl},
			Preference: 10,
			Mx:         record.CName,
		})
	} else {
		fmt.Printf("Recieved a request for unknow record type: %d\n", msg.Qtype)
	}
	return answers
}

func (s *RedisDNSServer) Lookup(msg dns.Question) *Record {
	log.Printf("LOOKUP: Looking for '%s'", msg.Name)
	str, err := s.redisClient.Get(msg.Name)
	log.Printf("Msg Name is '%s'", msg.Name)
	log.Printf("LOOKUP: Found %s; Err: %v", str, err)
	var result Record
	if err == nil {
		json.Unmarshal([]byte(str), &result)
	} else {
		log.Printf("No record for %s, trying wildcard %s", msg.Name, wildCardHostName(msg.Name))
		domainDots := strings.Count(s.domain, ".")
		if strings.Count(msg.Name, ".") > domainDots+1 {
			str, err := s.redisClient.Get(wildCardHostName(msg.Name))
			if err == nil {
				json.Unmarshal([]byte(str), &result)
			}
		}
	}
	return &result
}

func wildCardHostName(hostName string) string {
	nameParts := strings.SplitAfterN(hostName, ".", 2)
	return "*." + nameParts[1]
}

func (s *RedisDNSServer) SOA(msg dns.Question) dns.RR {
	return &dns.SOA{
		Hdr:     dns.RR_Header{Name: s.domain, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: 60},
		Ns:      s.hostname,
		Mbox:    s.mbox,
		Serial:  uint32(time.Now().Unix()),
		Refresh: 86400,
		Retry:   7200,
		Expire:  86400,
		Minttl:  60,
	}
}
