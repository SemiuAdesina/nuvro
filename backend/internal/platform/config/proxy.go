package config

import (
	"strings"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clientip"
)

type Proxy struct {
	TrustedHops  int
	TrustedCIDRs []string
}

func (r *reader) proxy() Proxy {
	return Proxy{
		TrustedHops:  r.number("NUVRO_TRUSTED_PROXY_HOPS", 0),
		TrustedCIDRs: r.list("NUVRO_TRUSTED_PROXY_CIDRS"),
	}
}

func (r *reader) list(key string) []string {
	raw := r.text(key, "")
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	values := make([]string, 0, strings.Count(raw, ",")+1)
	for _, entry := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(entry); trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}

func (p Proxy) Resolver() (*clientip.Resolver, error) {
	return clientip.NewResolver(p.TrustedCIDRs, p.TrustedHops)
}

func (c Config) proxyProblems() []string {
	if _, err := c.Proxy.Resolver(); err != nil {
		return []string{"NUVRO_TRUSTED_PROXY_HOPS and NUVRO_TRUSTED_PROXY_CIDRS are inconsistent: " + err.Error()}
	}
	if c.Proxy.TrustedHops > len(c.Proxy.TrustedCIDRs)+maximumProxyDepth {
		return []string{"NUVRO_TRUSTED_PROXY_HOPS is larger than any deployment this service supports"}
	}
	return nil
}

const maximumProxyDepth = 4
