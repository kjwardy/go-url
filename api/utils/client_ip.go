package utils

import (
	"net"
	"strings"
)

func ClientIP(remoteAddr, forwardedFor string, allowForwardedFor bool, trustLevel int) string {
	if allowForwardedFor && trustLevel > 0 && forwardedFor != "" {
		forwardedIPs := strings.Split(forwardedFor, ",")
		index := len(forwardedIPs) - trustLevel
		if index < 0 {
			index = 0
		}
		if ip := RemoteIP(strings.TrimSpace(forwardedIPs[index])); ip != "" {
			return ip
		}
	}
	return RemoteIP(remoteAddr)
}

func RemoteIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = strings.Trim(remoteAddr, "[]")
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.String()
	}
	return ""
}
