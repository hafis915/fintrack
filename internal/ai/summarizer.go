package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type CategoryStat struct {
	Name       string `json:"name"`
	Allocated  int64  `json:"allocated"`
	Spent      int64  `json:"spent"`
	Percentage int    `json:"percentage"`
}

type WeeklyStats struct {
	UserName        string         `json:"user_name,omitempty"`
	WeekStart       string         `json:"week_start"` // ISO date
	WeekEnd         string         `json:"week_end"`
	TotalSpent      int64          `json:"total_spent"`
	TotalBudget     int64          `json:"total_budget"`
	TopCategories   []CategoryStat `json:"top_categories"`
	VsLastWeekDelta int64          `json:"vs_last_week_delta"` // negative = spent less, positive = spent more
}

type Summarizer struct{ c *Client }

func NewSummarizer(c *Client) *Summarizer { return &Summarizer{c: c} }

// Narrative produces a 2-3 sentence weekly recap in Bahasa Indonesia,
// coach-style (per PRD §02 — friendly, non-judgmental, contextual).
func (s *Summarizer) Narrative(ctx context.Context, stats WeeklyStats) (string, error) {
	statsJSON, _ := json.Marshal(stats)
	system := "Kamu adalah financial coach untuk fresh worker Indonesia. Bicara ramah, singkat, dan tidak menghakimi. Selalu berikan satu saran konkret."
	prompt := fmt.Sprintf(`Buat ringkasan mingguan berdasarkan data berikut, MAKSIMUM 3 kalimat.

Data: %s

Format yang diharapkan:
1. Kalimat pertama: gambaran umum minggu ini (apakah on-track atau tidak).
2. Kalimat kedua: perbandingan dengan minggu sebelumnya (jika data tersedia).
3. Kalimat ketiga: satu saran konkret untuk minggu depan.

Tulis HANYA paragraf narasi, tanpa nomor, tanpa heading.`, string(statsJSON))

	raw, err := s.c.Complete(ctx, CompleteOptions{
		System: system,
		Messages: []Message{
			{Role: "user", Content: []Block{NewTextBlock(prompt)}},
		},
		MaxTokens:   300,
		Temperature: 0.4,
	})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(raw), nil
}
