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

// *.key.domain or *-key.domain
func wildcardKeys(lookupValue string) []string {
	keys := []string{}
	// *.key
	nameParts := strings.SplitAfterN(lookupValue, ".", 2)
	keys[0] = "*." + nameParts[1]
	// *-key
	keys[1] = "*-" + lookupValue
	return keys
}

// Get the record for key, apply wildcard if bare record doesn't work
func (s *RedisDNSServer) Get(key string) *Record {
	keys := append([]string{key}, wildcardKeys(key)...)
	r := &Record{}

	for i := 0; i < len(keys); i++ {
		bAry := Lookup(s.redisClient, keys[i])
		if bAry != nil {
			err := r.Parse(bAry)
			if err != nil {
				log.Printf("Error parsing JSON  %s: %s\n", keys[i], err)
				return r
			}
			if r.TTL == 0 {
				r.TTL = TTL
			}
			return r
		}
	}
	return r // Nothing was found
}

// Parse value from Redis into Record
func (r *Record) Parse(bAry []byte) error {
	err := json.Unmarshal(bAry, r)
	if err != nil {
		return err
	}
	return nil
}
