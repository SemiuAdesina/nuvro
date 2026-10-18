package identity

type Sealed struct {
	ID       string
	UserID   string
	Provider string
	Token    []byte
}

func (s Sealed) Revocable() bool { return len(s.Token) > 0 }
