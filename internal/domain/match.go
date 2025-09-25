package domain

import "time"

type MatchState string

const (
	MatchScheduled  MatchState = "scheduled"
	MatchCompleted  MatchState = "completed"
	MatchConflicted MatchState = "conflicted"
)

type ResultType string

const (
	P1Won ResultType = "p1"
	P2Won ResultType = "p2"
	Draw  ResultType = "draw"
)

type Match struct {
	ID           MatchID
	TournamentID TournamentID
	Round        Round

	P1 ParticipantID
	P2 ParticipantID

	State     MatchState
	OpinionP1 *ResultType
	OpinionP2 *ResultType
	Result    *ResultType

	ScheduledAt *time.Time
}
