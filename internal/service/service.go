package service

import (
	"errors"
	"math"
	"time"

	"github.com/Ycyken/tournament-bot/internal/domain"
	"github.com/Ycyken/tournament-bot/internal/store"
)

type Service struct {
	store store.Store
}

func New(s store.Store) *Service {
	return &Service{store: s}
}

func (s *Service) CreateTournament(owner domain.TelegramUserID, title string) (*domain.Tournament, error) {
	t := domain.NewTournament(owner, title, domain.Swiss)
	id, err := s.store.CreateTournament(t)
	if err != nil {
		return nil, err
	}

	t.ID = id

	return t, nil
}

func (s *Service) StartTournament(tid domain.TournamentID) error {
	t, err := s.store.GetTournament(tid)
	if err != nil {
		return err
	}

	if t.CurrentRound > 0 {
		return errors.New("tournament already started")
	}

	participantsCount := len(t.Participants)
	if participantsCount < 2 {
		return errors.New("need at least 2 participants to start tournament")
	}

	// ⌊log₂(N)⌋ + 1
	if participantsCount == 2 {
		t.LastRound = 1
	} else {
		t.LastRound = domain.Round(int(math.Log2(float64(len(t.Participants)))) + 1)
	}

	if err := t.DrawNewRound(); err != nil {
		return err
	}

	err = s.store.SaveTournament(t)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetUserTournaments(userID domain.TelegramUserID) (map[domain.TournamentID]string, error) {
	return s.store.GetUserTournaments(userID)
}

func (s *Service) ApplyToTournament(app *domain.Application) error {
	t, err := s.store.GetTournament(app.TournamentID)
	if err != nil {
		return err
	}

	if t.CurrentRound > 0 {
		return errors.New("tournament already started")
	}

	for _, p := range t.Participants {
		for _, id := range p.Roster {
			if id == app.TelegramUserID {
				return errors.New("user already in tournament")
			}
		}
	}

	appls, err := s.GetApplications(app.TournamentID)
	if err != nil {
		return err
	}
	for _, a := range appls {
		if a.TelegramUserID == app.TelegramUserID {
			return errors.New("user is already applied to this tournament")
		}
	}

	return s.store.CreateApplication(app)
}

func (s *Service) ApproveApplication(tID domain.TournamentID, tgID domain.TelegramUserID) error {
	t, err := s.store.GetTournament(tID)
	if err != nil {
		return err
	}

	if t.CurrentRound > 0 {
		return errors.New("tournament already started")
	}

	app, err := s.store.GetApplication(tID, tgID)
	if err != nil {
		return err
	}
	p := &domain.Participant{
		ID:           domain.ParticipantID(len(t.Participants)),
		Name:         app.Name,
		TelegramTag:  app.TelegramTag,
		TournamentID: app.TournamentID,
		Kind:         domain.ParticipantKindUser,
		Roster:       []domain.TelegramUserID{app.TelegramUserID},
		JoinedAt:     time.Now(),
	}

	t.Participants = append(t.Participants, p)

	if err := s.store.DeleteApplication(tID, tgID); err != nil {
		return err
	}
	if err := s.store.SaveTournament(t); err != nil {
		return err
	}

	return nil
}

func (s *Service) RejectApplication(tID domain.TournamentID, tgID domain.TelegramUserID) error {
	return s.store.DeleteApplication(tID, tgID)
}

func (s *Service) GetApplications(tid domain.TournamentID) ([]*domain.Application, error) {
	return s.store.GetApplications(tid)
}

func (s *Service) GetApplication(tid domain.TournamentID, uid domain.TelegramUserID) (*domain.Application, error) {
	return s.store.GetApplication(tid, uid)
}

func (s *Service) GetTournament(tid domain.TournamentID) (*domain.Tournament, error) {
	return s.store.GetTournament(tid)
}

func (s *Service) GetTournaments() ([]*domain.Tournament, error) {
	return s.store.GetTournaments()
}

func (s *Service) ReportMatchResult(tournamentID domain.TournamentID, matchID domain.MatchID, userID domain.TelegramUserID, result domain.ResultType) (*domain.Tournament, error) {
	t, err := s.store.GetTournament(tournamentID)
	if err != nil {
		return nil, err
	}

	var participantID domain.ParticipantID = -1
outerLoop:
	for _, p := range t.Participants {
		for _, id := range p.Roster {
			if id == userID {
				participantID = p.ID
				break outerLoop
			}
		}
	}

	if err := t.ReportOpinion(matchID, participantID, result); err != nil {
		return nil, err
	}

	if err := s.store.SaveTournament(t); err != nil {
		return nil, err
	}

	return t, nil
}

func (s *Service) SetMatchResultByAdmin(tournamentID domain.TournamentID, matchID domain.MatchID, adminID domain.TelegramUserID, result domain.ResultType) (*domain.Tournament, error) {
	t, err := s.store.GetTournament(tournamentID)
	if err != nil {
		return nil, err
	}

	if t.OwnerID != adminID {
		return nil, errors.New("only tournament owner can set match results")
	}

	if err := t.SetMatchResultByAdmin(matchID, result); err != nil {
		return nil, err
	}

	if err := s.store.SaveTournament(t); err != nil {
		return nil, err
	}

	return t, nil
}
