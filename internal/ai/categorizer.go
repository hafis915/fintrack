package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Category struct {
	ID   uuid.UUID
	Name string
	Type string
}

type ReceiptAlternative struct {
	CategoryName string  `json:"category_name"`
	Confidence   float64 `json:"confidence"`
}

type ReceiptScan struct {
	Amount       int64                `json:"amount"`
	CategoryName string               `json:"category_name"`
	Note         string               `json:"note"`
	Confidence   float64              `json:"confidence"`
	Alternatives []ReceiptAlternative `json:"alternatives"`
}

type Categorizer struct{ c *Client }

func NewCategorizer(c *Client) *Categorizer { return &Categorizer{c: c} }

// Scan sends the receipt image + the list of allowed categories to the model,
// expects JSON back, and resolves the chosen category name to its UUID.
// Returns uuid.Nil if the model picks a name that isn't in `available`.
func (cz *Categorizer) Scan(ctx context.Context, image []byte, mimeType string, available []Category) (*ReceiptScan, uuid.UUID, error) {
	if len(available) == 0 {
		return nil, uuid.Nil, errors.New("categorizer: no categories provided")
	}
	cats, _ := json.Marshal(catNames(available))
	system := "Kamu asisten pencatat keuangan untuk pengguna Indonesia. Jawab HANYA dengan JSON valid, tanpa teks tambahan, tanpa markdown."
	prompt := fmt.Sprintf(`Baca struk pembelian pada gambar. Kembalikan JSON dengan schema:
{
  "amount": int_rupiah_total,
  "category_name": "salah satu dari daftar di bawah",
  "note": "merchant atau item utama (maks 50 karakter)",
  "confidence": 0.0-1.0,
  "alternatives": [{"category_name":"...","confidence":...}]
}
Pilih category_name hanya dari daftar ini: %s.
Jika tidak yakin, pilih yang paling mungkin dan turunkan confidence.`, string(cats))

	raw, err := cz.c.Complete(ctx, CompleteOptions{
		System: system,
		Messages: []Message{
			{Role: "user", Content: []Block{
				NewImageBlock(mimeType, image),
				NewTextBlock(prompt),
			}},
		},
		MaxTokens:   600,
		Temperature: 0.1,
		JSONOnly:    true,
	})
	if err != nil {
		return nil, uuid.Nil, err
	}

	// Defensive: some models still wrap JSON in code fences despite being told not to.
	cleaned := stripCodeFence(raw)

	var out ReceiptScan
	if err := json.Unmarshal([]byte(cleaned), &out); err != nil {
		return nil, uuid.Nil, fmt.Errorf("categorizer: parse model JSON: %w (raw=%s)", err, truncate(cleaned, 200))
	}
	id := matchCategory(out.CategoryName, available)
	return &out, id, nil
}

func catNames(cats []Category) []string {
	out := make([]string, 0, len(cats))
	for _, c := range cats {
		out = append(out, c.Name)
	}
	return out
}

func matchCategory(name string, cats []Category) uuid.UUID {
	for _, c := range cats {
		if strings.EqualFold(c.Name, name) {
			return c.ID
		}
	}
	return uuid.Nil
}

func stripCodeFence(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		// drop first line (the fence + optional language)
		if i := strings.Index(s, "\n"); i >= 0 {
			s = s[i+1:]
		}
		s = strings.TrimSuffix(strings.TrimSuffix(strings.TrimSpace(s), "```"), "json")
	}
	return strings.TrimSpace(s)
}
