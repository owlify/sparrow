package utils

import (
	"fmt"
	"strings"
)

func ConvertToString(v interface{}) string {
	if v == nil {
		return ""
	}

	stringV := fmt.Sprintf("%v", v)
	return stringV
}

func StringContains(s string, strArray []string, caseSensitive bool) bool {
	for _, str := range strArray {
		if caseSensitive {
			if s == str {
				return true
			}
		} else {
			if strings.EqualFold(str, s) {
				return true
			}
		}
	}
	return false
}
