/*
Copyright AppsCode Inc. and Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package redis_finder

import "testing"

func TestParseAnnounceEndpoint(t *testing.T) {
	tests := []struct {
		name       string
		value      string
		host       string
		port       string
		busPort    string
		shouldFail bool
	}{
		{
			name:    "bare hostname",
			value:   "redis-node-s0-0.example.com",
			host:    "redis-node-s0-0.example.com",
			port:    "6379",
			busPort: "16379",
		},
		{
			name:    "documented endpoint",
			value:   "redis-node-s0-0.example.com:6380@16380",
			host:    "redis-node-s0-0.example.com",
			port:    "6380",
			busPort: "16380",
		},
		{
			name:    "bracketed ipv6 endpoint",
			value:   "[2001:db8::10]:6379@16379",
			host:    "2001:db8::10",
			port:    "6379",
			busPort: "16379",
		},
		{
			name:       "missing host",
			value:      ":6379@16379",
			shouldFail: true,
		},
		{
			name:       "missing bus port",
			value:      "redis.example.com:6379@",
			shouldFail: true,
		},
		{
			name:       "multiple bus separators",
			value:      "redis.example.com:6379@16379@extra",
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, port, busPort, err := parseAnnounceEndpoint(tt.value)
			if tt.shouldFail {
				if err == nil {
					t.Fatalf("parseAnnounceEndpoint(%q) succeeded, want error", tt.value)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseAnnounceEndpoint(%q) returned error: %v", tt.value, err)
			}
			if host != tt.host || port != tt.port || busPort != tt.busPort {
				t.Fatalf("parseAnnounceEndpoint(%q) = %q %q %q, want %q %q %q", tt.value, host, port, busPort, tt.host, tt.port, tt.busPort)
			}
		})
	}
}
