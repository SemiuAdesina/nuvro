package config

import "net/url"

func (c Config) connectionProblems() []string {
	var problems []string
	for _, setting := range []struct {
		name, value string
		schemes     []string
	}{
		{"NUVRO_DATABASE_URL", c.Database.URL, []string{"postgres", "postgresql"}},
		{"NUVRO_REDIS_URL", c.Redis.URL, []string{"redis", "rediss"}},
	} {
		if setting.value == "" {
			continue
		}
		parsed, err := url.Parse(setting.value)
		if err != nil || parsed.Hostname() == "" {
			problems = append(problems, setting.name+" must be a valid connection URL")
			continue
		}
		accepted := false
		for _, scheme := range setting.schemes {
			if parsed.Scheme == scheme {
				accepted = true
			}
		}
		if !accepted {
			problems = append(problems, setting.name+" uses an unsupported scheme")
		}
	}
	return problems
}
