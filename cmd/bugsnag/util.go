package main

import "strings"

func splitByEquals(strs []string) map[string]string {
	m := map[string]string{}
	for _, kvp := range strs {
		if kvp != "" {
			pair := strings.SplitN(kvp, "=", 2)
			m[pair[0]] = pair[1]
		}
	}
	return m
}
