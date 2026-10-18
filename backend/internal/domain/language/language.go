package language

import (
	"errors"
	"regexp"
	"strings"
)

var ErrInvalid = errors.New("language: invalid tag")

var tagPattern = regexp.MustCompile(`^[a-z]{2,3}(-[A-Z][a-z]{3})?(-([A-Z]{2}|[0-9]{3}))?$`)

type Code string

const Undetermined Code = "und"

func Parse(raw string) (Code, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", ErrInvalid
	}
	parts := strings.Split(trimmed, "-")
	parts[0] = strings.ToLower(parts[0])
	if len(parts) > 1 && len(parts[1]) == 4 {
		parts[1] = strings.ToUpper(parts[1][:1]) + strings.ToLower(parts[1][1:])
	}
	if last := len(parts) - 1; last > 0 && len(parts[last]) == 2 {
		parts[last] = strings.ToUpper(parts[last])
	}
	candidate := Code(strings.Join(parts, "-"))
	if !candidate.Valid() {
		return "", ErrInvalid
	}
	return candidate, nil
}

func (c Code) Valid() bool { return tagPattern.MatchString(string(c)) }

func (c Code) Base() Code {
	base, _, _ := strings.Cut(string(c), "-")
	return Code(base)
}

func (c Code) String() string { return string(c) }

func (c Code) SameLanguage(other Code) bool { return c.Base() == other.Base() }

var languageByRegion = map[string]Code{
	"PT": "pt-PT", "BR": "pt-BR", "ES": "es-ES", "MX": "es-MX", "CO": "es-CO",
	"FR": "fr-FR", "IT": "it-IT", "DE": "de-DE", "NL": "nl-NL", "GR": "el-GR",
	"JP": "ja-JP", "KR": "ko-KR", "TH": "th-TH", "ID": "id-ID", "VN": "vi-VN",
	"TR": "tr-TR", "MA": "ar-MA", "AE": "ar-AE", "TZ": "sw-TZ", "ZA": "en-ZA",
	"IS": "is-IS", "NZ": "en-NZ", "BB": "en-BB", "NG": "en-NG", "US": "en-US", "GB": "en-GB",
}

func ForRegion(region string) (Code, bool) {
	code, found := languageByRegion[strings.ToUpper(strings.TrimSpace(region))]
	return code, found
}
