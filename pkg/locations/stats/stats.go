// Package stats tracks private-location proxy and DNS activity between usage reports.
package stats

import "sync/atomic"

var requests atomic.Uint64
var dnsRequests atomic.Uint64

// IncRequest records one proxied request.
func IncRequest() {
	requests.Add(1)
}

// IncDNS records one forwarded DNS query.
func IncDNS() {
	dnsRequests.Add(1)
}

// Snapshot holds activity since the previous snapshot.
type Snapshot struct {
	Requests uint64
	DNS      uint64
}

// SnapshotAndReset returns counts since the last snapshot and clears them.
func SnapshotAndReset() Snapshot {
	return Snapshot{
		Requests: requests.Swap(0),
		DNS:      dnsRequests.Swap(0),
	}
}
