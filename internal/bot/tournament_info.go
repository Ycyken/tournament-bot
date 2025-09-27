package bot

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Ycyken/tournament-bot/internal/domain"
)

func TournamentInfoMessage(t *domain.Tournament) string {
	var text strings.Builder
	fmt.Fprintf(&text, "🏆 Турнир %s (ID %d)\n", t.Title, t.ID)
	fmt.Fprintf(&text, "Количество раундов: %d\n", t.LastRound)
	fmt.Fprintf(&text, "Текущий раунд: %d\n\n", t.CurrentRound)
	fmt.Fprintf(&text, "Участники:\n")
	sort.Slice(t.Participants, func(i, j int) bool {
		return t.Participants[i].Score > t.Participants[j].Score
	})
	for _, p := range t.Participants {
		tag := ""
		if p.TelegramTag != nil {
			tag = " (@" + *p.TelegramTag + ")"
		}
		fmt.Fprintf(&text, "%s%s: %.1f\n", p.Name, tag, p.Score)
	}

	return text.String()
}
