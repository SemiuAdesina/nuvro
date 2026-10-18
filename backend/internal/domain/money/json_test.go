package money_test

import (
	"encoding/json"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
)

func TestAmountSurvivesAJSONRoundTrip(t *testing.T) {
	original, err := money.New(500_000_000, money.Currency("NGN"))
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded money.Amount
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Minor() != original.Minor() || decoded.Currency() != original.Currency() {
		t.Fatalf("round trip lost data: %s became %s", original, decoded)
	}
}

func TestAmountJSONShapeIsStable(t *testing.T) {
	amount, err := money.New(2_500, money.Currency("EUR"))
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(amount)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != `{"amountMinor":2500,"currency":"EUR"}` {
		t.Fatalf("encoded = %s", encoded)
	}
}

func TestZeroValueAmountEncodesAsNull(t *testing.T) {
	encoded, err := json.Marshal(money.Amount{})
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != "null" {
		t.Fatalf("encoded = %s", encoded)
	}
	var decoded money.Amount
	if err := json.Unmarshal([]byte("null"), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Currency() != "" {
		t.Fatal("null decoded into a currency")
	}
}

func TestMalformedAmountsAreRejected(t *testing.T) {
	for name, raw := range map[string]string{
		"not an object":    `"5 EUR"`,
		"no currency":      `{"amountMinor":100}`,
		"invalid currency": `{"amountMinor":100,"currency":"eur"}`,
		"long currency":    `{"amountMinor":100,"currency":"EUROS"}`,
	} {
		var decoded money.Amount
		if err := json.Unmarshal([]byte(raw), &decoded); err == nil {
			t.Fatalf("%s was accepted", name)
		}
	}
}

func TestNestedAmountsRoundTrip(t *testing.T) {
	type holder struct {
		Price money.Amount `json:"price"`
	}
	amount, err := money.New(22_000, money.Currency("EUR"))
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(holder{Price: amount})
	if err != nil {
		t.Fatal(err)
	}
	var decoded holder
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Price.Minor() != 22_000 || decoded.Price.Currency().String() != "EUR" {
		t.Fatalf("nested round trip lost data: %+v", decoded.Price)
	}
}
