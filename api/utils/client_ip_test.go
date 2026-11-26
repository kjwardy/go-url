package utils

import "testing"

func TestClientIP(t *testing.T) {
	tests := []struct {
		name              string
		remoteAddr        string
		forwardedFor      string
		allowForwardedFor bool
		trustLevel        int
		want              string
	}{
		{
			name:         "ignores forwarded header unless enabled",
			remoteAddr:   "198.51.100.10:1323",
			forwardedFor: "203.0.113.8",
			trustLevel:   1,
			want:         "198.51.100.10",
		},
		{
			name:              "uses forwarded client address when enabled",
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
			name:              "falls back to peer for invalid forwarded address",
			remoteAddr:        "198.51.100.10:1323",
			forwardedFor:      "not-an-ip",
			allowForwardedFor: true,
			trustLevel:        1,
			want:              "198.51.100.10",
		},
		{
			name:       "parses IPv6 peer",
			remoteAddr: "[2001:db8::1]:1323",
			want:       "2001:db8::1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ClientIP(test.remoteAddr, test.forwardedFor, test.allowForwardedFor, test.trustLevel)
			if got != test.want {
				t.Errorf("ClientIP() = %q, want %q", got, test.want)
			}
		})
	}
}
