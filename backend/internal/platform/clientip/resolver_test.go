package clientip_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clientip"
)

const loadBalancer = "10.0.0.0/8"
const contentNetwork = "192.0.2.0/24"

func resolver(t *testing.T, hops int, ranges ...string) *clientip.Resolver {
	t.Helper()
	value, err := clientip.NewResolver(ranges, hops)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestADirectClientIsReadFromTheConnection(t *testing.T) {
	direct := resolver(t, 0)
	for name, test := range map[string]struct{ remote, want string }{
		"ipv4":         {"203.0.113.9:51000", "203.0.113.9"},
		"ipv6":         {"[2001:db8::5]:51000", "2001:db8::5"},
		"no port":      {"203.0.113.9", "203.0.113.9"},
		"mapped ipv4":  {"[::ffff:203.0.113.9]:443", "203.0.113.9"},
		"zoned ipv6":   {"[fe80::1%eth0]:443", "fe80::1"},
		"unparseable":  {"not-an-address", clientip.Unknown},
		"empty remote": {"", clientip.Unknown},
	} {
		if got := direct.Resolve(test.remote, nil); got != test.want {
			t.Fatalf("%s: %q resolved to %q, want %q", name, test.remote, got, test.want)
		}
	}
}

func TestADirectClientIgnoresAnyForwardingItSends(t *testing.T) {
	direct := resolver(t, 0)
	spoofed := []string{"198.51.100.7, 203.0.113.200"}
	if got := direct.Resolve("203.0.113.9:51000", spoofed); got != "203.0.113.9" {
		t.Fatalf("a direct client spoofed its address to %q", got)
	}
}

func TestOneTrustedProxyYieldsTheClient(t *testing.T) {
	behind := resolver(t, 1, loadBalancer)
	for name, test := range map[string]struct {
		remote    string
		forwarded []string
		want      string
	}{
		"ipv4 client":  {"10.0.0.4:443", []string{"203.0.113.9"}, "203.0.113.9"},
		"ipv6 client":  {"10.0.0.4:443", []string{"2001:db8::5"}, "2001:db8::5"},
		"ipv6 proxy":   {"[::ffff:10.0.0.4]:443", []string{"203.0.113.9"}, "203.0.113.9"},
		"spaced":       {"10.0.0.4:443", []string{"  203.0.113.9  "}, "203.0.113.9"},
		"split header": {"10.0.0.4:443", []string{"203.0.113.9"}, "203.0.113.9"},
	} {
		if got := behind.Resolve(test.remote, test.forwarded); got != test.want {
			t.Fatalf("%s resolved to %q, want %q", name, got, test.want)
		}
	}
}

func TestATrustedChainWalksBackByTheConfiguredHops(t *testing.T) {
	chained := resolver(t, 2, loadBalancer, contentNetwork)
	forwarded := []string{"203.0.113.9, 192.0.2.10"}
	if got := chained.Resolve("10.0.0.4:443", forwarded); got != "203.0.113.9" {
		t.Fatalf("a CDN and load balancer chain resolved to %q", got)
	}
	split := []string{"203.0.113.9", "192.0.2.10"}
	if got := chained.Resolve("10.0.0.4:443", split); got != "203.0.113.9" {
		t.Fatalf("a chain split across headers resolved to %q", got)
	}
}

func TestSpoofedEntriesToTheLeftCannotShiftTheClient(t *testing.T) {
	chained := resolver(t, 2, loadBalancer, contentNetwork)
	spoofed := []string{"1.1.1.1, 2.2.2.2, 203.0.113.9, 192.0.2.10"}
	if got := chained.Resolve("10.0.0.4:443", spoofed); got != "203.0.113.9" {
		t.Fatalf("prepended entries moved the client to %q", got)
	}
	behind := resolver(t, 1, loadBalancer)
	if got := behind.Resolve("10.0.0.4:443", []string{"9.9.9.9, 203.0.113.9"}); got != "203.0.113.9" {
		t.Fatalf("a spoofed leading entry produced %q", got)
	}
}

func TestAnUntrustedPeerIsNeverAllowedToForward(t *testing.T) {
	behind := resolver(t, 1, loadBalancer)
	if got := behind.Resolve("203.0.113.50:51000", []string{"198.51.100.7"}); got != "203.0.113.50" {
		t.Fatalf("an untrusted peer forwarded an address: %q", got)
	}
	if got := behind.Resolve("[2001:db8::99]:443", []string{"198.51.100.7"}); got != "2001:db8::99" {
		t.Fatalf("an untrusted IPv6 peer forwarded an address: %q", got)
	}
}

func TestMalformedChainsFallBackToThePeer(t *testing.T) {
	chained := resolver(t, 2, loadBalancer)
	for name, forwarded := range map[string][]string{
		"too short":     {"203.0.113.9"},
		"empty":         {""},
		"empty entry":   {"203.0.113.9, , 192.0.2.10"},
		"not addresses": {"nonsense, rubbish"},
		"bad at index":  {"203.0.113.9, not-an-ip, 192.0.2.10"},
		"missing":       nil,
		"oversized":     {strings.Repeat("203.0.113.9, ", 40) + "192.0.2.10"},
		"too many":      {strings.TrimSuffix(strings.Repeat("1.1.1.1,", 40), ",")},
	} {
		if got := chained.Resolve("10.0.0.4:443", forwarded); got != "10.0.0.4" {
			t.Fatalf("%s resolved to %q, want the trusted peer", name, got)
		}
	}
}

func TestTheChosenEntryMustItselfBeAnAddress(t *testing.T) {
	behind := resolver(t, 1, loadBalancer)
	if got := behind.Resolve("10.0.0.4:443", []string{"203.0.113.9, not-an-ip"}); got != "10.0.0.4" {
		t.Fatalf("a malformed chosen entry resolved to %q", got)
	}
}

func TestConfigurationRefusesHopsWithoutTrustedRanges(t *testing.T) {
	if _, err := clientip.NewResolver(nil, 1); !errors.Is(err, clientip.ErrHopsWithoutTrust) {
		t.Fatalf("err = %v, want a refusal to trust forwarding with no trusted ranges", err)
	}
	if _, err := clientip.NewResolver([]string{"10.0.0.0/8"}, -1); err == nil {
		t.Fatal("a negative hop count was accepted")
	}
	if _, err := clientip.NewResolver([]string{"not-a-cidr"}, 1); err == nil {
		t.Fatal("an invalid CIDR was accepted")
	}
	if _, err := clientip.NewResolver([]string{"10.0.0.4/8", "  ", ""}, 1); err != nil {
		t.Fatalf("an unmasked CIDR with blank entries was refused: %v", err)
	}
}

func TestANilResolverBehavesLikeADirectDeployment(t *testing.T) {
	var absent *clientip.Resolver
	if got := absent.Resolve("203.0.113.9:51000", []string{"198.51.100.7"}); got != "203.0.113.9" {
		t.Fatalf("a nil resolver honoured forwarding: %q", got)
	}
}

func FuzzResolveNeverTrustsAnUntrustedPeer(f *testing.F) {
	f.Add("203.0.113.9:51000", "198.51.100.7")
	f.Add("10.0.0.4:443", "203.0.113.9, 192.0.2.10")
	f.Add("[2001:db8::5]:443", "")
	f.Add("nonsense", "1.1.1.1")
	behind, err := clientip.NewResolver([]string{loadBalancer}, 1)
	if err != nil {
		f.Fatal(err)
	}
	f.Fuzz(func(t *testing.T, remote, forwarded string) {
		resolved := behind.Resolve(remote, []string{forwarded})
		if resolved == "" {
			t.Fatal("resolution produced an empty address")
		}
		if resolved == clientip.Unknown {
			return
		}
		if strings.ContainsAny(resolved, " ,\t\r\n") {
			t.Fatalf("resolved %q carries chain punctuation", resolved)
		}
		direct, directErr := clientip.NewResolver(nil, 0)
		if directErr != nil {
			t.Fatal(directErr)
		}
		peer := direct.Resolve(remote, nil)
		if peer != clientip.Unknown && !strings.HasPrefix(peer, "10.") && resolved != peer {
			t.Fatalf("an untrusted peer %q was resolved as %q", peer, resolved)
		}
	})
}
