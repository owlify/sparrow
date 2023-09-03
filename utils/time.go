package utils

import (
	"time"
)

func DurationByUnit(v int, unit string) time.Duration {
	if unit == "s" {
		return time.Duration(v) * time.Second
	} else if unit == "m" {
		return time.Duration(v) * time.Minute
	} else if unit == "h" {
		return time.Duration(v) * time.Hour
	}
	return time.Duration(v)
}
