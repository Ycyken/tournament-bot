package bot

import (
	"fmt"
	"strings"

	"github.com/Ycyken/tournament-bot/internal/domain"
)

func TournamentInfoMessage(t *domain.Tournament) string {
	var text strings.Builder
	fmt.Fprintf(&text, "🏆 Турнир %s (ID %d)\n", t.Title, t.ID)
	fmt.Fprintf(&text, "Количество раундов: %d\n", t.LastRound)
	fmt.Fprintf(&text, "Текущий раунд: %d\n\n", t.CurrentRound)
	fmt.Fprintf(&text, "Участники:\n")
	for _, p := range t.Participants {
		tag := ""
		if p.TelegramTag != nil {
			tag = " (@" + *p.TelegramTag + ")"
		}
		fmt.Fprintf(&text, "%s%s: %.1f\n", p.Name, tag, p.Score)
	}

	return text.String()
}
