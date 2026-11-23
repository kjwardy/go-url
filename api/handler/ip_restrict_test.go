package handler

import "testing"

func TestIpAllowed(t *testing.T) {
	tables := []struct {
		allowedIps []string
		actualIps  []string
		allowed    bool
	}{
		{[]string{"10.1.1.1"}, []string{"10.1.1.1"}, true},             // same IP
		{[]string{"10.1.1.1"}, []string{"10.1.1.0"}, false},            // different IP
		{[]string{"10.1.1.1"}, []string{"10.1.1.0", "10.1.1.1"}, true}, // multiple IPs
		{[]string{"10.1.1.1", "10.1.1.2"}, []string{"10.1.1.2"}, true}, // multiple allowedIPs
		{[]string{"10.1.1.1/32"}, []string{"10.1.1.1"}, true},          // allowed CIDR
		{[]string{"10.1.1.1/32"}, []string{"10.1.1.0"}, false},         // disallowed CIDR
		{[]string{"10.1.1.1/28"}, []string{"10.1.1.0"}, true},          // first matching ip CIDR
		{[]string{"10.1.1.1/28"}, []string{"10.1.1.15"}, true},         // last matching ip CIDR
		{[]string{"10.1.1.1/28"}, []string{"10.1.1.16"}, false},        // first non matching matching ip CIDR
		{[]string{"10.1.1.1/28"}, []string{"10.1.0.255"}, false},       // last non matching matching ip CIDR
	}

	for _, table := range tables {
		allowed := ipAllowed(table.allowedIps, table.actualIps)
		if allowed != table.allowed {
			t.Errorf("Expected %t, with allowedIps: %q and actualIps: %q", table.allowed, table.allowedIps, table.actualIps)
		}
	}
}

func TestSourceIPTrustConfiguration(t *testing.T) {
	tests := []struct {
		name              string
		remoteAddr        string
		forwardedFor      string
		allowForwardedFor bool
		trustLevel        int
		want              string
	}{
		{
			name:              "ignores forwarded header unless enabled",
			remoteAddr:        "198.51.100.10:1323",
			forwardedFor:      "203.0.113.8",
			allowForwardedFor: false,
			trustLevel:        1,
			want:              "198.51.100.10",
		},
		{
			name:              "uses client address through one trusted proxy",
			remoteAddr:        "198.51.100.10:1323",
			forwardedFor:      "203.0.113.8",
			allowForwardedFor: true,
			trustLevel:        1,
			want:              "203.0.113.8",
		},
		{
			name:              "selects client address using trusted proxy count",
			remoteAddr:        "198.51.100.10:1323",
			forwardedFor:      "203.0.113.8, 198.51.100.9",
			allowForwardedFor: true,
			trustLevel:        2,
			want:              "203.0.113.8",
		},
		{
			name:              "falls back to direct address for invalid forwarded value",
			remoteAddr:        "198.51.100.10:1323",
			forwardedFor:      "not-an-ip",
			allowForwardedFor: true,
			trustLevel:        1,
			want:              "198.51.100.10",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := sourceIP(test.remoteAddr, test.forwardedFor, test.allowForwardedFor, test.trustLevel)
			if got != test.want {
				t.Errorf("sourceIP() = %q, want %q", got, test.want)
			}
		})
	}
}
