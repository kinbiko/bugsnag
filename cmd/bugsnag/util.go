package main

import (
	"log"
	"strings"
)

func splitByEquals(strs []string) map[string]string {
	kvps := map[string]string{}
	for _, kvp := range strs {
		if kvp != "" {
			pair := strings.SplitN(kvp, "=", 2)
			kvps[pair[0]] = pair[1]
		}
	}
	return kvps
}

func logf(msg string, args ...interface{}) {
	log.Printf(msg, args...)
}
