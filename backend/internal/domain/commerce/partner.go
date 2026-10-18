package commerce

import (
	"errors"
	"net/url"
	"slices"
	"strings"
)

var ErrInvalid = errors.New("commerce: invalid record")
var ErrUntrustedDestination = errors.New("commerce: destination is not an approved partner domain")

type ProductType string

const (
	Flight    ProductType = "flight"
	Stay      ProductType = "stay"
	Activity  ProductType = "activity"
	Transfer  ProductType = "transfer"
	Car       ProductType = "car"
	ESIM      ProductType = "esim"
	Insurance ProductType = "insurance"
)

func ProductTypes() []ProductType {
	return []ProductType{Flight, Stay, Activity, Transfer, Car, ESIM, Insurance}
}

func (p ProductType) Valid() bool { return slices.Contains(ProductTypes(), p) }

type PartnerStatus string

const (
	Enabled  PartnerStatus = "enabled"
	Disabled PartnerStatus = "disabled"
	Degraded PartnerStatus = "degraded"
)

func (s PartnerStatus) Valid() bool {
	switch s {
	case Enabled, Disabled, Degraded:
		return true
	default:
		return false
	}
}

func (s PartnerStatus) Usable() bool { return s == Enabled || s == Degraded }

type Partner struct {
	Code            string
	DisplayName     string
	Status          PartnerStatus
	Products        []ProductType
	AllowedDomains  []string
	AttributionText string
}

func (p Partner) Validate() error {
	if p.Code == "" || p.DisplayName == "" || !p.Status.Valid() {
		return ErrInvalid
	}
	if len(p.Products) == 0 || len(p.AllowedDomains) == 0 {
		return ErrInvalid
	}
	for _, product := range p.Products {
		if !product.Valid() {
			return ErrInvalid
		}
	}
	for _, domain := range p.AllowedDomains {
		if domain == "" || strings.ContainsAny(domain, "/: ") {
			return ErrInvalid
		}
	}
	return nil
}

func (p Partner) Supports(product ProductType) bool { return slices.Contains(p.Products, product) }

func (p Partner) AllowsDestination(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return ErrUntrustedDestination
	}
	if parsed.User != nil {
		return ErrUntrustedDestination
	}
	host := strings.ToLower(parsed.Hostname())
	for _, allowed := range p.AllowedDomains {
		candidate := strings.ToLower(allowed)
		if host == candidate || strings.HasSuffix(host, "."+candidate) {
			return nil
		}
	}
	return ErrUntrustedDestination
}
