package domain

import (
	"errors"
	"math"
	"sort"
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
	LastRound    Round

	Matches      map[Round][]*Match
	Participants []*Participant
	Opponents    map[ParticipantID]map[ParticipantID]bool
	Byes         map[ParticipantID]bool

	StartTime time.Time
}

func NewTournament(ownerID TelegramUserID, title string, system System) *Tournament {
	return &Tournament{
		OwnerID:      ownerID,
		Title:        title,
		System:       system,
		CurrentRound: 0,
		LastRound:    0,

		Matches:      make(map[Round][]*Match),
		Participants: []*Participant{},
		Opponents:    make(map[ParticipantID]map[ParticipantID]bool),
		Byes:         make(map[ParticipantID]bool),

		StartTime: time.Now(),
	}
}

func (t *Tournament) UserParticipates(tgID TelegramUserID) bool {
	for _, p := range t.Participants {
		for _, uid := range p.Roster {
			if uid == tgID {
				return true
			}
		}
	}
	return false
}

func (t *Tournament) ReportOpinion(matchID MatchID, pID ParticipantID, result ResultType) error {
	match := t.Matches[t.CurrentRound][matchID]
	if match == nil {
		return ErrMatchNotFound
	}

	switch pID {
	case match.P1:
		match.OpinionP1 = &result
	case match.P2:
		match.OpinionP2 = &result
	default:
		return ErrParticipantNotInMatch
	}

	if match.OpinionP1 != nil && match.OpinionP2 != nil {
		if *match.OpinionP1 == *match.OpinionP2 {
			match.Result = match.OpinionP1
			match.State = MatchCompleted
		} else {
			match.State = MatchConflicted
		}
	}

	return nil
}

func (t *Tournament) SetMatchResultByAdmin(matchID MatchID, result ResultType) error {
	match := t.Matches[t.CurrentRound][matchID]
	if match == nil {
		return ErrMatchNotFound
	}

	match.Result = &result
	match.State = MatchCompleted
	return nil
}

func (t *Tournament) DrawNewRound() error {
	for _, m := range t.Matches[t.CurrentRound] {
		if m.State != MatchCompleted {
			return ErrNotAllMatchesCompleted
		}
	}

	for _, m := range t.Matches[t.CurrentRound] {
		switch *m.Result {
		case P1Won:
			t.Participants[m.P1].Score += 1.0
		case P2Won:
			t.Participants[m.P2].Score += 1.0
		case Draw:
			t.Participants[m.P1].Score += 0.5
			t.Participants[m.P2].Score += 0.5
		}
	}
	if t.CurrentRound == t.LastRound {
		return nil
	}
	t.CurrentRound++

	pNumber := len(t.Participants)
	var pairing []ParticipantID
	if t.CurrentRound == 1 {
		pairing = shuffledIDs(pNumber)
	} else {
		pairing = pairParticipantsSwiss(t)
	}
	for i := 0; i < len(pairing)-1; i += 2 {
		t.Matches[t.CurrentRound] = append(t.Matches[t.CurrentRound], &Match{
			ID:           MatchID(len(t.Matches[t.CurrentRound]) + 1),
			TournamentID: t.ID,
			Round:        t.CurrentRound,
			P1:           pairing[i],
			P2:           pairing[i+1],
			State:        MatchScheduled,
		})
	}
	if pNumber%2 == 1 {
		t.Participants[pairing[pNumber-1]].Score += 1.0
		t.Byes[pairing[pNumber-1]] = true
	}

	return nil
}

func pairParticipantsSwiss(t *Tournament) []ParticipantID {
	var byeID ParticipantID = -1
	var lowestScore float64 = math.MaxFloat64
	for _, p := range t.Participants {
		if !t.Byes[p.ID] && (p.Score < lowestScore) {
			byeID = p.ID
			lowestScore = p.Score
		}
	}
	if len(t.Participants)%2 == 0 {
		byeID = -1
	}

	participantsCopy := make([]Participant, len(t.Participants))
	for i, p := range t.Participants {
		participantsCopy[i] = *p
	}
	sort.Slice(participantsCopy, func(i, j int) bool {
		return participantsCopy[i].Score > participantsCopy[j].Score
	})

	var pairing []ParticipantID
	paired := make(map[ParticipantID]bool)

	for i := 0; i < len(participantsCopy); i++ {
		p1 := participantsCopy[i].ID

		for j := i + 1; j < len(participantsCopy); j++ {
			p2 := participantsCopy[j].ID
			if p1 == byeID || p2 == byeID {
				continue
			}
			if !t.Opponents[p1][p2] && !paired[p2] {
				pairing = append(pairing, p1, p2)
				paired[p1] = true
				paired[p2] = true
				break
			}
		}
	}

	if len(t.Participants)%2 == 1 {
		pairing = append(pairing, byeID)
	}

	return pairing
}
