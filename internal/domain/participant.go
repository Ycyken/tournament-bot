package domain

import "time"

type ParticipantKind string

const (
	ParticipantKindUser ParticipantKind = "user"
	ParticipantKindTeam ParticipantKind = "team"
)

type Participant struct {
	ID           ParticipantID
	Name         string
	TournamentID TournamentID
	Kind         ParticipantKind
	Roster       []TelegramUserID

	Eliminated bool
	Score      float64 // for Swiss
	JoinedAt   time.Time
}
