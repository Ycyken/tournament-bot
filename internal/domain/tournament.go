package domain

import (
	"errors"
	"time"
)

var (
	ErrMatchNotFound          = errors.New("match not found")
	ErrParticipantNotInMatch  = errors.New("participant is not in match")
	ErrNotAllMatchesCompleted = errors.New("not all matches in current round are completed")
	ErrSingleElimination      = errors.New("single elimination is not implemented yet")
)

type System string
type Round int

const (
	SingleElimination System = "single_elimination"
	Swiss             System = "swiss"
)

type Tournament struct {
	ID           TournamentID
	OwnerID      TelegramUserID
	Title        string
	System       System
	CurrentRound Round
	TotalRounds  Round

	Matches      map[Round]map[MatchID]*Match
	Participants map[ParticipantID]*Participant
	opponents    map[ParticipantID]map[ParticipantID]struct{}
	byes         map[ParticipantID]struct{}

	StartTime time.Time
}
