package config

import "time"

type Runtime struct {
	AlertInterval      time.Duration
	PriceWatchInterval time.Duration
	RetentionInterval  time.Duration
}

func (r *reader) runtime() Runtime {
	return Runtime{
		AlertInterval:      r.duration("NUVRO_ALERT_INTERVAL", 5*time.Minute),
		PriceWatchInterval: r.duration("NUVRO_PRICE_WATCH_INTERVAL", 15*time.Minute),
		RetentionInterval:  r.duration("NUVRO_RETENTION_INTERVAL", 6*time.Hour),
	}
}

func (c Config) runtimeProblems() []string {
	var problems []string
	if c.Runtime.AlertInterval < time.Minute || c.Runtime.AlertInterval > time.Hour {
		problems = append(problems, "NUVRO_ALERT_INTERVAL must be between 1m and 1h")
	}
	if c.Runtime.PriceWatchInterval < time.Minute || c.Runtime.PriceWatchInterval > 6*time.Hour {
		problems = append(problems, "NUVRO_PRICE_WATCH_INTERVAL must be between 1m and 6h")
	}
	if c.Runtime.RetentionInterval < time.Minute || c.Runtime.RetentionInterval > 24*time.Hour {
		problems = append(problems, "NUVRO_RETENTION_INTERVAL must be between 1m and 24h")
	}
	return problems
}
