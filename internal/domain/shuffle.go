package domain

import (
	"math/rand"
	"time"
)

func shuffledIDs(n int) []ParticipantID {
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))

	ids := make([]ParticipantID, n)
	for i := 0; i < n; i++ {
		ids[i] = ParticipantID(i)
	}

	rng.Shuffle(len(ids), func(i, j int) {
		ids[i], ids[j] = ids[j], ids[i]
	})

	return ids
}
