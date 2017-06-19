package main

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/hoisie/redis"
)

// Lookup the record in Redis
func Lookup(client redis.Client, key string) []byte {
	log.Printf("⌘ Lookup %s\n", key)
	byteary, err := client.Get(key)
	log.Printf("  %s\n", string(byteary))
	if err != nil {
		log.Printf("Error during Lookup %s: %s\n", key, err)
		return nil
	}
	return byteary
}

// WildCardLookup of the record in Redis
func WildCardLookup(client redis.Client, key string) []byte {
	wc := wildcardHostName(key)
	return Lookup(client, wc)
}

func wildcardHostName(hostName string) string {
	nameParts := strings.SplitAfterN(hostName, ".", 2)
	return "*." + nameParts[1]
}

// Get the record for key, apply wildcard if bare record doesn't work
func (s *RedisDNSServer) Get(key string) *Record {
	r := &Record{}
	bAry := Lookup(s.redisClient, key)
	if bAry == nil {
		bAry = WildCardLookup(s.redisClient, key)
		if bAry == nil {
			return r
		}
	}
	err := r.Parse(bAry)
	if err != nil {
		log.Printf("Error parsing JSON  %s: %s\n", key, err)
		return r
	}
	if r.TTL == 0 {
		r.TTL = TTL
	}
	return r
}

// Parse value from Redis into Record
func (r *Record) Parse(bAry []byte) error {
	err := json.Unmarshal(bAry, r)
	if err != nil {
		return err
	}
	return nil
}
