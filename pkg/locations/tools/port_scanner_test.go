package tools

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParsePortScanRequest(t *testing.T) {
	payloadJSON := `{"target": "escape.tech", "ports": "1-10000", "rate": 20, "mode": "c", "timeout": 20000}`

	var payload map[string]any
	if err := json.NewDecoder(strings.NewReader(payloadJSON)).Decode(&payload); err != nil {
		t.Fatalf("decode request body: %v", err)
	}

	req, err := parsePortScanRequest(payload)
	if err != nil {
		t.Fatalf("parse port scan request: %v", err)
	}

	if req.Timeout != 20000 {
		t.Errorf("expected timeout to be 1000, got %d", req.Timeout)
	}

	if req.Rate != 20 {
		t.Errorf("expected rate to be 100, got %d", req.Rate)
	}
}

func TestParsePortScanRequestAllowsMissingPorts(t *testing.T) {
	payloadJSON := `{"target": "escape.tech", "rate": 20, "mode": "c", "timeout": 20000}`

	var payload map[string]any
	if err := json.NewDecoder(strings.NewReader(payloadJSON)).Decode(&payload); err != nil {
		t.Fatalf("decode request body: %v", err)
	}

	req, err := parsePortScanRequest(payload)
	if err != nil {
		t.Fatalf("parse port scan request: %v", err)
	}

	if req.Ports != "" {
		t.Fatalf("expected ports to be empty when omitted, got %q", req.Ports)
	}
}

func TestParsePortScanRequestPortList(t *testing.T) {
	payloadJSON := `{"target": "escape.tech", "port_list": "shodan", "rate": 20, "mode": "c", "timeout": 20000}`

	var payload map[string]any
	if err := json.NewDecoder(strings.NewReader(payloadJSON)).Decode(&payload); err != nil {
		t.Fatalf("decode request body: %v", err)
	}

	req, err := parsePortScanRequest(payload)
	if err != nil {
		t.Fatalf("parse port scan request: %v", err)
	}

	if req.Ports != shodanPorts {
		t.Fatalf("expected shodan ports list")
	}
}

func TestParsePortScanRequestRejectsPortsAndPortList(t *testing.T) {
	payloadJSON := `{"target": "escape.tech", "ports": "80", "port_list": "shodan", "rate": 20, "mode": "c", "timeout": 20000}`

	var payload map[string]any
	if err := json.NewDecoder(strings.NewReader(payloadJSON)).Decode(&payload); err != nil {
		t.Fatalf("decode request body: %v", err)
	}

	_, err := parsePortScanRequest(payload)
	if err == nil || err.Error() != "ports and port_list are mutually exclusive" {
		t.Fatalf("expected mutually exclusive error, got %v", err)
	}
}

func TestParsePortScanRequestRejectsUnknownPortList(t *testing.T) {
	payloadJSON := `{"target": "escape.tech", "port_list": "top-ports", "rate": 20, "mode": "c", "timeout": 20000}`

	var payload map[string]any
	if err := json.NewDecoder(strings.NewReader(payloadJSON)).Decode(&payload); err != nil {
		t.Fatalf("decode request body: %v", err)
	}

	_, err := parsePortScanRequest(payload)
	if err == nil {
		t.Fatal("expected unsupported port_list error")
	}
}
