package postgres

import (
	"database/sql"
	"errors"

	"github.com/Ycyken/tournament-bot/internal/domain"
)

func (s *PostgresStore) CreateApplication(app *domain.Application) error {
	_, err := s.db.Exec(`
			INSERT INTO applications (tournament_id, telegram_user_id, name, telegram_tag, text)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT DO NOTHING
		`, app.TournamentID, app.TelegramUserID, app.Name, app.TelegramTag, app.Text)
	return err
}

func (s *PostgresStore) GetApplications(tournamentID domain.TournamentID) ([]*domain.Application, error) {
	rows, err := s.db.Query(`
		SELECT tournament_id, telegram_user_id, name, telegram_tag, text
		FROM applications WHERE tournament_id = $1
	`, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []*domain.Application
	for rows.Next() {
		app := &domain.Application{}
		if err := rows.Scan(&app.TournamentID, &app.TelegramUserID, &app.Name, &app.TelegramTag, &app.Text); err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}
	return apps, nil
}

func (s *PostgresStore) GetApplication(tID domain.TournamentID, tgID domain.TelegramUserID) (*domain.Application, error) {
	row := s.db.QueryRow(`
		SELECT tournament_id, telegram_user_id, name, telegram_tag, text
		FROM applications
		WHERE tournament_id = $1 AND telegram_user_id = $2
	`, tID, tgID)

	app := &domain.Application{}
	if err := row.Scan(&app.TournamentID, &app.TelegramUserID, &app.Name, &app.TelegramTag, &app.Text); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return app, nil
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
