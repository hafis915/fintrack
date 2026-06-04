package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// stubClient is a deterministic, network-free Client used by tests and by local
// dev runs where no OPEN_ROUTER_API_KEY is configured. It performs lightweight
// regex NLU on the LAST user message: it looks for a rupiah amount and a known
// flexible-category name substring, then emits the same STRICT JSON contract a
// real model would (per the planner system prompt). It NEVER invents budget
// numbers beyond echoing what the user explicitly asked for — the deterministic
// re-balance still happens in app code.
type stubClient struct{}

// NewStubClient returns a deterministic Client. It is selected when
// OPEN_ROUTER_API_KEY is empty. It never errors and never touches the network.
func NewStubClient() Client {
	return stubClient{}
}

// stubFlexibleCategories are the flexible ('want'/'variable') category names the
// stub can recognise in free text. Matching is case-insensitive substring. These
// mirror the flexible categories the planner suggests so the stub can stand in
// for a real model in tests and offline dev.
var stubFlexibleCategories = []string{
	"makan & minum",
	"makan",
	"belanja",
	"transportasi",
	"transport",
	"kesehatan",
	"hiburan",
	"nongkrong",
	"self-care",
	"self care",
}

// rupiahAmount matches a number written as digits, optionally with thousands
// separators (.,space) and an optional juta/rb/ribu/jt suffix:
//
//	"1500000", "1.500.000", "1,5jt", "1.5 juta", "750rb", "750 ribu"
var rupiahAmount = regexp.MustCompile(`(?i)(?:rp\.?\s*)?([\d][\d.,\s]*)\s*(jt|juta|rb|ribu|k)?`)

// stubAdjustment mirrors one entry of the STRICT JSON "adjustments" array the
// planner handler parses. category_name is matched against the live flexible
// categories; target_amount is the rupiah integer the user asked for.
type stubAdjustment struct {
	CategoryName string `json:"category_name"`
	TargetAmount int64  `json:"target_amount"`
}

type stubReply struct {
	Reply       string           `json:"reply"`
	Adjustments []stubAdjustment `json:"adjustments"`
}

func (stubClient) Complete(_ context.Context, _ string, messages []Message) (string, error) {
	last := lastUserMessage(messages)

	cat := matchCategory(last)
	amt, okAmt := matchAmount(last)

	if cat == "" || !okAmt {
		out, _ := json.Marshal(stubReply{
			Reply:       "Boleh — kategori mana yang mau diubah?",
			Adjustments: []stubAdjustment{},
		})
		return string(out), nil
	}

	out, _ := json.Marshal(stubReply{
		Reply: fmt.Sprintf("Oke, aku set %s jadi Rp %d dan seimbangkan yang lain.", cat, amt),
		Adjustments: []stubAdjustment{
			{CategoryName: cat, TargetAmount: amt},
		},
	})
	return string(out), nil
}

// lastUserMessage returns the content of the most recent message with role
// "user", or "" if there is none.
func lastUserMessage(messages []Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			return messages[i].Content
		}
	}
	return ""
}

// matchCategory returns the first known flexible-category name found as a
// case-insensitive substring of s, or "" if none. Longer names are listed first
// in stubFlexibleCategories so "makan & minum" wins over bare "makan".
func matchCategory(s string) string {
	low := strings.ToLower(s)
	for _, c := range stubFlexibleCategories {
		if strings.Contains(low, c) {
			return c
		}
	}
	return ""
}

// matchAmount extracts a rupiah amount from s, expanding jt/juta/rb/ribu/k
// suffixes. Returns (amount, true) on success, (0, false) if no number found.
func matchAmount(s string) (int64, bool) {
	m := rupiahAmount.FindStringSubmatch(s)
	if m == nil {
		return 0, false
	}

	digits := m[1]
	suffix := strings.ToLower(m[2])

	// Decimal forms like "1,5jt" / "1.5 juta" only make sense with a multiplier
	// suffix. With a suffix, treat the last separator as a decimal point.
	if suffix != "" {
		f, ok := parseDecimal(digits)
		if !ok {
			return 0, false
		}
		switch suffix {
		case "jt", "juta":
			return int64(f * 1_000_000), true
		case "rb", "ribu", "k":
			return int64(f * 1_000), true
		}
	}

	// No suffix: strip all separators and parse as a plain integer.
	clean := stripSeparators(digits)
	if clean == "" {
		return 0, false
	}
	n, err := strconv.ParseInt(clean, 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

// parseDecimal parses a possibly-separated number like "1", "1,5", "1.5",
// "1.500" into a float. The LAST separator is treated as the decimal point; any
// earlier separators are thousands and stripped.
func parseDecimal(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	lastDot := strings.LastIndexByte(s, '.')
	lastComma := strings.LastIndexByte(s, ',')
	sep := lastDot
	if lastComma > sep {
		sep = lastComma
	}
	if sep < 0 {
		f, err := strconv.ParseFloat(stripSeparators(s), 64)
		return f, err == nil
	}
	intPart := stripSeparators(s[:sep])
	fracPart := stripSeparators(s[sep+1:])
	if intPart == "" {
		intPart = "0"
	}
	f, err := strconv.ParseFloat(intPart+"."+fracPart, 64)
	return f, err == nil
}

// stripSeparators removes '.', ',', and whitespace, leaving only digits.
func stripSeparators(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
