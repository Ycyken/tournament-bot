package domain

import (
	"errors"
	"fmt"
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
	p := t.FindParticipantBytgID(tgID)
	if p == nil {
		return false
	}
	return true
}

func (t *Tournament) FindCurrentMatch(pID ParticipantID) *Match {
	for _, m := range t.Matches[t.CurrentRound] {
		if m.P1 == pID || m.P2 == pID {
			return m
		}
	}
	return nil
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

	_ = t.DrawNewRound()

	return nil
}

func (t *Tournament) SetMatchResultByAdmin(matchID MatchID, result ResultType) error {
	match := t.Matches[t.CurrentRound][matchID]
	if match == nil {
		return ErrMatchNotFound
	}

	match.Result = &result
	match.State = MatchCompleted
	_ = t.DrawNewRound()
	return nil
}

func (t *Tournament) FindParticipantByPID(pID ParticipantID) *Participant {
	for _, p := range t.Participants {
		if p.ID == pID {
			return p
		}
	}
	return nil
}

func (t *Tournament) FindParticipantBytgID(tgID TelegramUserID) *Participant {
	for _, p := range t.Participants {
		for _, id := range p.Roster {
			if id == tgID {
				return p
			}
		}
	}
	return nil
}

func (t *Tournament) GetParticipantMatches(pID ParticipantID) []*Match {
	var matches []*Match
	for r := Round(1); r <= t.CurrentRound; r++ {
		for _, m := range t.Matches[r] {
			if m.P1 == pID || m.P2 == pID {
				matches = append(matches, m)
				break
			}
		}
	}
	return matches
}

func (t *Tournament) GetMatchesHistory(pID ParticipantID) string {
	matches := t.GetParticipantMatches(pID)

	var text string
	for _, m := range matches {
		if m == nil {
			text += fmt.Sprintf("Раунд %d: у вас пока не назначен матч.\n\n", m.Round)
			break
		}
		p := t.FindParticipantByPID(pID)
		var tag string
		if p.TelegramTag != nil {
			tag = "(@" + *p.TelegramTag + ")"
		}
		text += fmt.Sprintf("Раунд %d: матч против %s%s:\n", m.Round, p.Name, tag)
		text += fmt.Sprintf("Состояние: %s\n", m.State)
		if m.Result != nil {
			if m.P1 == pID && *m.Result == P1Won ||
				m.P2 == pID && *m.Result == P2Won {
				text += "Результат: Вы выиграли!"
			} else if *m.Result == "draw" {
				text += "Результат: Ничья!"
			} else {
				text += "Результат: Вы проиграли."
			}
		} else {
			text += "Матч еще не завершен."
		}
		text += "\n\n"
	}

	return text
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
			ID:           t.nextMatchID(),
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

func (t *Tournament) nextMatchID() MatchID {
	var matchID int
	for _, ms := range t.Matches {
		matchID += len(ms)
	}
	return MatchID(matchID)
}

func pairParticipantsSwiss(t *Tournament) []ParticipantID {
	var byeID ParticipantID = -1
	var lowestScore = math.MaxFloat64
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
