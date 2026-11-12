package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/projectdiscovery/goflags"
	"github.com/projectdiscovery/naabu/v2/pkg/result"
	"github.com/projectdiscovery/naabu/v2/pkg/runner"
)

const (
	defaultPortScanRate    = 1000
	defaultPortScanMode    = "c"
	defaultPortScanTimeout = 60 * time.Second
	defaultScanRetries     = 3
	portScanTool           = "port-scanner"
)

type portScanRequest struct {
	Timeout   int    `json:"timeout"` // in seconds
	Rate      int    `json:"rate"`
	Ports     string `json:"ports"`
	Mode      string `json:"mode"`
	Target    string `json:"target"`
	MaxThread int    `json:"max_thread,omitempty"`
}

type portScanResults struct {
	Hosts map[string][]int `json:"hosts"`
	Ports []int            `json:"ports"`
}

func buildPortScannerTool() tool {
	return tool{
		name:             portScanTool,
		executionTimeout: portScanTimeout,
		executor:         executePortScan,
	}
}

func portScanTimeout(payload map[string]any) (time.Duration, error) {
	req, err := parsePortScanRequest(payload)
	if err != nil {
		return 0, err
	}

	return time.Duration(req.Timeout) * time.Second, nil
}

func createNaabuOptions(req portScanRequest, onResult result.ResultFn, onReceive result.ResultFn) runner.Options {
	options := runner.Options{
		Host:      goflags.StringSlice{req.Target},
		Ports:     req.Ports,
		ScanType:  req.Mode,
		Rate:      req.Rate,
		Timeout:   time.Second,
		Retries:   defaultScanRetries,
		Silent:    true,
		OnResult:  onResult,
		OnReceive: onReceive,
	}
	if req.MaxThread > 0 {
		options.Threads = req.MaxThread
	}

	return options
}

func executePortScan(ctx context.Context, payload map[string]any, progress func(jobResult)) (jobResult, error) {
	req, err := parsePortScanRequest(payload)
	if err != nil {
		return nil, err
	}

	results := portScanResults{Hosts: map[string][]int{}}
	var mu sync.Mutex
	options := createNaabuOptions(req,
		// Result callback
		func(hr *result.HostResult) {
			mu.Lock()
			defer mu.Unlock()
			addPorts(results.Hosts, hr)
		},
		// Receive callback
		func(hr *result.HostResult) {
			mu.Lock()
			defer mu.Unlock()
			addPorts(results.Hosts, hr)
			progress(portScanResultPayload(results))
		},
	)

	scanRunner, err := runner.NewRunner(&options)
	if err != nil {
		return portScanResultPayload(results), fmt.Errorf("failed to create port scan runner: %w", err)
	}

	defer func() { _ = scanRunner.Close() }()

	if err := scanRunner.RunEnumeration(ctx); err != nil {
		return portScanResultPayload(results), fmt.Errorf("port scan failed: %w", err)
	}

	return portScanResultPayload(results), nil
}

func parsePortScanRequest(payload map[string]any) (portScanRequest, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return portScanRequest{}, fmt.Errorf("failed to marshal payload: %w", err)
	}

	var req portScanRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return portScanRequest{}, fmt.Errorf("failed to marshal payload: %w", err)
	}

	if req.Rate <= 0 {
		return portScanRequest{}, errors.New("rate must be positive")
	}

	if req.Timeout <= 0 {
		return portScanRequest{}, errors.New("timeout must be positive")
	}

	if req.MaxThread < 0 {
		return portScanRequest{}, errors.New("max_thread must be non-negative")
	}

	return req, nil
}

func addPorts(hosts map[string][]int, hr *result.HostResult) {
	host := hr.Host
	if host == "" {
		host = hr.IP
	}

	if host == "" {
		return
	}

	seen := map[int]struct{}{}
	for _, port := range hosts[host] {
		seen[port] = struct{}{}
	}

	for _, port := range hr.Ports {
		if _, ok := seen[port.Port]; ok {
			continue
		}

		hosts[host] = append(hosts[host], port.Port)
		seen[port.Port] = struct{}{}
	}

	sort.Ints(hosts[host])
}

func portScanResultPayload(results portScanResults) jobResult {
	portsSeen := map[int]struct{}{}
	hosts := make(map[string][]int, len(results.Hosts))
	for host, ports := range results.Hosts {
		copied := append([]int(nil), ports...)
		sort.Ints(copied)
		hosts[host] = copied
		for _, port := range copied {
			portsSeen[port] = struct{}{}
		}
	}

	ports := make([]int, 0, len(portsSeen))
	for port := range portsSeen {
		ports = append(ports, port)
	}

	sort.Ints(ports)

	return jobResult{"hosts": hosts, "ports": ports}
}
