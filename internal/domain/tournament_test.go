package domain

import "testing"

func TestTournament_DrawNewRound(t *testing.T) {
	tourn := Tournament{
		Matches:   make(map[Round][]*Match),
		Opponents: make(map[ParticipantID]map[ParticipantID]bool),
		Byes:      make(map[ParticipantID]bool),
		LastRound: 5,
	}

	for i := 0; i < 15; i++ {
		tourn.Participants = append(tourn.Participants, &Participant{ID: ParticipantID(i), Roster: []TelegramUserID{TelegramUserID(i)}})
	}

	if err := tourn.DrawNewRound(); err != nil {
		t.Fatalf("DrawNewRound() error = %v", err)
	}

	if len(tourn.Matches[1]) != 7 {
		t.Fatalf("expected 8 matches in round 1, got %d", len(tourn.Matches[1]))
	}

}
