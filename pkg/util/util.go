package util

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseGoVersion(version string) (major, minor, patch int, err error) {
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return 0, 0, 0, fmt.Errorf("invalid go version: %s", version)
	}
	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return
	}
	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return
	}
	if len(parts) > 2 {
		patch, err = strconv.Atoi(parts[2])
		if err != nil {
			return
		}
	}
	return
}
