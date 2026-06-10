//go:build windows

package dns

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// systemResolvers enumerates the DNS servers configured on every up adapter
// via GetAdaptersAddresses. Windows has no /etc/resolv.conf, so this is the
// equivalent discovery path. Returned addresses are bare IPs; upstreams.go
// applies the port and link-local filtering.
func systemResolvers() ([]string, error) {
	const flags = windows.GAA_FLAG_SKIP_ANYCAST | windows.GAA_FLAG_SKIP_MULTICAST

	var size uint32
	// First call sizes the buffer; a nil pointer with ERROR_BUFFER_OVERFLOW is expected.
	_ = windows.GetAdaptersAddresses(syscall.AF_UNSPEC, flags, 0, nil, &size)
	if size == 0 {
		return nil, nil
	}

	buf := make([]byte, size)
	addrs := (*windows.IpAdapterAddresses)(unsafe.Pointer(&buf[0]))
	if err := windows.GetAdaptersAddresses(syscall.AF_UNSPEC, flags, 0, addrs, &size); err != nil {
		return nil, err
	}

	var out []string
	for a := addrs; a != nil; a = a.Next {
		if a.OperStatus != windows.IfOperStatusUp {
			continue
		}
		for d := a.FirstDnsServerAddress; d != nil; d = d.Next {
			if ip := d.Address.IP(); ip != nil {
				out = append(out, ip.String())
			}
		}
	}
	return out, nil
}
