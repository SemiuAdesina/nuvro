package clientip

import (
	"errors"
	"net"
	"net/netip"
	"strings"
)

const ForwardedHeader = "X-Forwarded-For"
const maximumChainEntries = 32
const maximumChainLength = 1024
const Unknown = "unknown"

var ErrHopsWithoutTrust = errors.New("clientip: trusted hops require trusted proxy ranges")

type Resolver struct {
	trusted []netip.Prefix
	hops    int
}

func NewResolver(ranges []string, hops int) (*Resolver, error) {
	if hops < 0 {
		return nil, errors.New("clientip: trusted hops must not be negative")
	}
	trusted := make([]netip.Prefix, 0, len(ranges))
	for _, raw := range ranges {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			return nil, errors.New("clientip: " + value + " is not a CIDR range")
		}
		trusted = append(trusted, prefix.Masked())
	}
	if hops > 0 && len(trusted) == 0 {
		return nil, ErrHopsWithoutTrust
	}
	return &Resolver{trusted: trusted, hops: hops}, nil
}

func (r *Resolver) Resolve(remoteAddr string, forwarded []string) string {
	peer, ok := parse(hostOf(remoteAddr))
	if !ok {
		return Unknown
	}
	if r == nil || r.hops == 0 || !r.trusts(peer) {
		return peer.String()
	}
	client, found := r.walk(forwarded)
	if !found {
		return peer.String()
	}
	return client.String()
}

func (r *Resolver) trusts(address netip.Addr) bool {
	for _, prefix := range r.trusted {
		if prefix.Contains(address) {
			return true
		}
	}
	return false
}

func (r *Resolver) walk(forwarded []string) (netip.Addr, bool) {
	chain, ok := chainOf(forwarded)
	if !ok || len(chain) < r.hops {
		return netip.Addr{}, false
	}
	return parse(chain[len(chain)-r.hops])
}

func chainOf(forwarded []string) ([]string, bool) {
	joined := strings.Join(forwarded, ",")
	if len(joined) > maximumChainLength {
		return nil, false
	}
	entries := make([]string, 0, maximumChainEntries)
	for _, raw := range strings.Split(joined, ",") {
		value := strings.TrimSpace(raw)
		if value == "" {
			return nil, false
		}
		if len(entries) == maximumChainEntries {
			return nil, false
		}
		entries = append(entries, value)
	}
	return entries, true
}

func hostOf(remoteAddr string) string {
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return host
	}
	return remoteAddr
}

func parse(raw string) (netip.Addr, bool) {
	value := strings.TrimSpace(raw)
	if strings.HasPrefix(value, "[") {
		value = hostOf(value)
	}
	address, err := netip.ParseAddr(value)
	if err != nil {
		if host, _, splitErr := net.SplitHostPort(value); splitErr == nil {
			address, err = netip.ParseAddr(host)
		}
		if err != nil {
			return netip.Addr{}, false
		}
	}
	return address.Unmap().WithZone(""), true
}
