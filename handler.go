package main

import (
	"log"

	"github.com/miekg/dns"
)

func (s *RedisDNSServer) handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.RecursionAvailable = true

	for _, msg := range r.Question {
		log.Printf("%v %s", dns.TypeToString[msg.Qtype], msg.Name)

		answers := s.Answer(msg)

		if len(answers) > 0 {
			m.Answer = append(m.Answer, answers...)
		} else {
			log.Printf("Warning, No answers\n")
		}
	}

	err := w.WriteMsg(m)
	if err != nil {
		log.Printf("Error writing msg %s\n", err)
	}
}
