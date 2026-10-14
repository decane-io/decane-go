package decane

import (
	"bytes"
	"encoding/json"
	"strconv"
)

// Amount is a decimal money value kept as a string so no digits are lost.
// The API sends amounts as strings in most places and as numbers in a few;
// both decode into an Amount.
type Amount string

// UnmarshalJSON accepts a JSON string or number.
func (a *Amount) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if string(b) == "null" {
		*a = ""
		return nil
	}
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*a = Amount(s)
		return nil
	}
	*a = Amount(string(b))
	return nil
}

// String returns the decimal text.
func (a Amount) String() string { return string(a) }

// Float64 parses the amount. Use it for display or comparison, never to send money.
func (a Amount) Float64() (float64, error) { return strconv.ParseFloat(string(a), 64) }

// FormatAmount turns a float64 into a plain decimal string ("12.5", never "1.25e+01").
// Prefer passing amounts as strings; this exists for values that start as floats.
func FormatAmount(f float64) string { return strconv.FormatFloat(f, 'f', -1, 64) }
