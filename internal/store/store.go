package store

import "github.com/Ycyken/tournament-bot/internal/domain"

type Store interface {
	Init() error
	CreateTournament(t *domain.Tournament) (domain.TournamentID, error)
	SaveTournament(t *domain.Tournament) error
	GetTournament(id domain.TournamentID) (*domain.Tournament, error)

	CreateApplication(app *domain.Application) error
	GetApplications(tournamentID domain.TournamentID) ([]*domain.Application, error)
	DeleteApplication(tournamentID domain.TournamentID, userID domain.TelegramUserID) error
}
