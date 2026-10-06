package private

import "testing"

func TestResolveEndpoint(t *testing.T) {
	tests := []struct {
		name      string
		transport string
		override  string
		want      transport
		wantAddr  string
		wantURL   string
		wantErr   bool
	}{
		{name: "default is ssh", want: transportSSH, wantAddr: "private-location.escape.tech:2222"},
		{name: "ssh override", transport: "ssh", override: "custom.example:2222", want: transportSSH, wantAddr: "custom.example:2222"},
		{name: "case and spaces", transport: " SSH ", want: transportSSH, wantAddr: "private-location.escape.tech:2222"},
		{
			name: "wss default", transport: "wss", want: transportWSS,
			wantAddr: "location.escape.tech:443", wantURL: "https://location.escape.tech/v2/connect",
		},
		{
			name: "https default", transport: "https", want: transportHTTPS,
			wantAddr: "location.escape.tech:443", wantURL: "https://location.escape.tech/v2/connect",
		},
		{
			name: "auto keeps only the host of a bare override", transport: "auto", override: "custom.example:2222", want: transportAuto,
			wantAddr: "custom.example:443", wantURL: "https://custom.example/v2/connect",
		},
		{
			name: "scheme override is used as is", transport: "https", override: "http://127.0.0.1:8080", want: transportHTTPS,
			wantAddr: "127.0.0.1:8080", wantURL: "http://127.0.0.1:8080/v2/connect",
		},
		{
			name: "wss scheme override with path", transport: "wss", override: "wss://edge.example:8443/custom", want: transportWSS,
			wantAddr: "edge.example:8443", wantURL: "https://edge.example:8443/custom",
		},
		{name: "unknown transport", transport: "quic", wantErr: true},
		{name: "unknown override scheme", transport: "https", override: "ftp://edge.example", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("ESCAPE_TRANSPORT", tt.transport)
			t.Setenv("ESCAPE_PRIVATE_LOCATION_URL", tt.override)

			ep, err := resolveEndpoint()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("resolveEndpoint() = %+v, want error", ep)
				}

				return
			}

			if err != nil {
				t.Fatalf("resolveEndpoint() error = %v", err)
			}

			if ep.transport != tt.want || ep.addr != tt.wantAddr {
				t.Fatalf("resolveEndpoint() = (%s, %s), want (%s, %s)", ep.transport, ep.addr, tt.want, tt.wantAddr)
			}

			gotURL := ""
			if ep.connectURL != nil {
				gotURL = ep.connectURL.String()
			}

			if gotURL != tt.wantURL {
				t.Fatalf("connectURL = %q, want %q", gotURL, tt.wantURL)
			}
		})
	}
}
