package evidence

type Attribution struct {
	Text     string
	URI      string
	Provider string
}

func (a Attribution) Valid() bool { return a.Text != "" && a.Provider != "" }
