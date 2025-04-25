package util

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func GetGoProxy() (string, error) {
	proxy := "https://proxy.golang.org"
	cmd := exec.Command("go", "env", "GOPROXY")
	out, err := cmd.Output()
	if err != nil {
		return proxy, nil
	}
	proxyEnv := strings.TrimSpace(string(out))
	if proxyEnv == "" {
		return proxy, nil
	}
	return firstGoProxy(proxyEnv), nil
}

func firstGoProxy(s string) string {
	if i := strings.Index(s, ","); i > 0 {
		return s[:i]
	}
	if i := strings.Index(s, "|"); i > 0 {
		return s[:i]
	}
	return s
}

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
