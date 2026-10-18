package admission

import "context"

type Class string

const (
	StandardRead    Class = "standard_read"
	ExpensiveSearch Class = "expensive_search"
	Authentication  Class = "authentication"
	CommerceWrite   Class = "commerce_write"
	Exempt          Class = ""
)

func (c Class) Valid() bool {
	switch c {
	case StandardRead, ExpensiveSearch, Authentication, CommerceWrite:
		return true
	default:
		return false
	}
}

type Allowance struct {
	Burst     int
	PerSecond float64
}

func (a Allowance) Valid() bool { return a.Burst > 0 && a.PerSecond > 0 }

type Limiter interface {
	Allow(context.Context, Class, string) (bool, error)
}

func Defaults() map[Class]Allowance {
	return map[Class]Allowance{
		StandardRead:    {Burst: 120, PerSecond: 4},
		ExpensiveSearch: {Burst: 12, PerSecond: 0.2},
		Authentication:  {Burst: 300, PerSecond: 10},
		CommerceWrite:   {Burst: 20, PerSecond: 0.5},
	}
}
