package domain

import "time"

type (
	TelegramUserID int64
	ParticipantID  int64
	TournamentID   int64
	MatchID        int64
)

type Tournament struct {
	ID           TournamentID
	OwnerID      TelegramUserID
	Title        string
	Participants []ParticipantID

	StartTime time.Time
}

type ParticipantKind string

const (
	ParticipantKindUser ParticipantKind = "user"
	ParticipantKindTeam ParticipantKind = "team"
)

type Participant struct {
	ID           ParticipantID
	TournamentID TournamentID
	Kind         ParticipantKind
	Roster       []TelegramUserID

	Eliminated bool
	Score      float64 // for Swiss
	JoinedAt   time.Time
}

type MatchState string

const (
	MatchScheduled MatchState = "scheduled"
	MatchCompleted MatchState = "completed"
)

type ResultType string

const (
	ResultP1   ResultType = "p1"
	ResultP2   ResultType = "p2"
	ResultDraw ResultType = "draw"
)

type Match struct {
	ID           MatchID
	TournamentID TournamentID

	P1 ParticipantID
	P2 ParticipantID

	State  MatchState
	Result *ResultType

	ScheduledAt *time.Time
}
