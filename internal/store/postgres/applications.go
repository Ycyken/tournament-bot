package postgres

import (
	"errors"
	"github.com/Ycyken/tournament-bot/internal/domain"
)

func (s *PostgresStore) CreateApplication(app *domain.Application) error {
	_, err := s.db.Exec(`
			INSERT INTO applications (tournament_id, telegram_user_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, app.TournamentID, app.TelegramUserID)
	return err
}

func (s *PostgresStore) GetApplications(tournamentID domain.TournamentID) ([]*domain.Application, error) {
	rows, err := s.db.Query(`
		SELECT tournament_id, telegram_user_id
		FROM applications WHERE tournament_id = $1
	`, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []*domain.Application
	for rows.Next() {
		app := &domain.Application{}
		if err := rows.Scan(&app.TournamentID, &app.TelegramUserID); err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}
	return apps, nil
}

func (s *PostgresStore) DeleteApplication(tournamentID domain.TournamentID, userID domain.TelegramUserID) error {
	res, err := s.db.Exec(`
		DELETE FROM applications
		WHERE tournament_id = $1 AND telegram_user_id = $2
	`, tournamentID, userID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return errors.New("application not found")
	}
	return nil
}
