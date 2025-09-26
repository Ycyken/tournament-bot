package postgres

import (
	"database/sql"
	"fmt"
	"github.com/Ycyken/tournament-bot/internal/domain"
)

func SaveTournamentParticipants(tx *sql.Tx, t *domain.Tournament) error {

	for _, p := range t.Participants {
		_, err := tx.Exec(`
			INSERT INTO participants (id, tournament_id, kind, name, eliminated, score, joined_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO UPDATE
			    SET eliminated = EXCLUDED.eliminated,
			        score = EXCLUDED.score
		`, p.ID, t.ID, p.Kind, fmt.Sprintf("P%d", p.ID), p.Eliminated, p.Score, p.JoinedAt)
		if err != nil {
			return err
		}

		// save telegram ids (rosters)
		for _, uid := range p.Roster {
			_, err := tx.Exec(`
				INSERT INTO participant_members (tournament_id, participant_id, telegram_user_id)
				VALUES ($1, $2, $3)
				ON CONFLICT DO NOTHING
			`, t.ID, p.ID, uid)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
