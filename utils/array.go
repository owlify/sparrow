package utils

import (
	"strings"
)

func ExistsInArray(array []string, slug string) bool {
	for _, s := range array {
		if strings.EqualFold(s, slug) {
			return true
		}
	}
	return false
}
