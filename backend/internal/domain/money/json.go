package money

import (
	"encoding/json"
	"errors"
)

var ErrMalformedJSON = errors.New("money: malformed amount")

type amountJSON struct {
	AmountMinor int64  `json:"amountMinor"`
	Currency    string `json:"currency"`
}

func (a Amount) MarshalJSON() ([]byte, error) {
	if a.currency == "" {
		return json.Marshal(nil)
	}
	return json.Marshal(amountJSON{AmountMinor: a.minor, Currency: a.currency.String()})
}

func (a *Amount) UnmarshalJSON(raw []byte) error {
	if string(raw) == "null" {
		*a = Amount{}
		return nil
	}
	var decoded amountJSON
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return ErrMalformedJSON
	}
	currency, err := ParseCurrency(decoded.Currency)
	if err != nil {
		return ErrMalformedJSON
	}
	value, err := New(decoded.AmountMinor, currency)
	if err != nil {
		return ErrMalformedJSON
	}
	*a = value
	return nil
}
