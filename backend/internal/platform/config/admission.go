package config

import (
	"strconv"
	"strings"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/admission"
)

type Admission struct {
	Enabled    bool
	Allowances map[admission.Class]admission.Allowance
}

func (r *reader) admission() Admission {
	value := Admission{Enabled: r.text("NUVRO_RATE_LIMIT_ENABLED", "true") != "false"}
	value.Allowances = map[admission.Class]admission.Allowance{}
	for class, fallback := range admission.Defaults() {
		name := "NUVRO_RATE_LIMIT_" + strings.ToUpper(string(class))
		value.Allowances[class] = admission.Allowance{
			Burst:     r.number(name+"_BURST", fallback.Burst),
			PerSecond: r.rate(name+"_PER_SECOND", fallback.PerSecond),
		}
	}
	return value
}

func (r *reader) rate(key string, fallback float64) float64 {
	raw := r.text(key, "")
	if raw == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		r.problems = append(r.problems, key+" must be a number of requests per second")
		return fallback
	}
	return parsed
}

func (c Config) admissionProblems() []string {
	var problems []string
	if !c.Admission.Enabled && c.RequiresBackingServices() {
		problems = append(problems, "NUVRO_RATE_LIMIT_ENABLED must not be false outside development")
	}
	for class, allowance := range c.Admission.Allowances {
		if !allowance.Valid() {
			problems = append(problems, "NUVRO_RATE_LIMIT_"+strings.ToUpper(string(class))+
				"_BURST and _PER_SECOND must both be positive")
		}
	}
	return problems
}
