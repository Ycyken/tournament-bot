package postgres

import (
	"database/sql"

	"github.com/Ycyken/tournament-bot/internal/domain"
)

func (s *PostgresStore) CreateTournament(t *domain.Tournament) (domain.TournamentID, error) {
	var id int64
	err := s.db.QueryRow(`
		INSERT INTO tournaments (owner_id, title, system, current_round, last_round, start_time)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, t.OwnerID, t.Title, t.System, t.CurrentRound, t.LastRound, t.StartTime).Scan(&id)
	if err != nil {
		return 0, err
	}
	return domain.TournamentID(id), nil
}

func (s *PostgresStore) SaveTournament(t *domain.Tournament) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// save round info
	_, err = tx.Exec(`
		UPDATE tournaments
		SET current_round = $1, last_round = $2
		WHERE id = $3
	`, t.CurrentRound, t.LastRound, t.ID)
	if err != nil {
		return err
	}

	// save participants
	err = SaveTournamentParticipants(tx, t)
	if err != nil {
		return err
	}

	// save opponents history
	for p1, opps := range t.Opponents {
		for p2 := range opps {
			_, err := tx.Exec(`
				INSERT INTO participant_opponents (tournament_id, participant1_id, participant2_id)
				VALUES ($1, $2, $3)
				ON CONFLICT DO NOTHING
			`, t.ID, p1, p2)
			if err != nil {
				return err
			}
		}
	}

	// save byes
	for pid := range t.Byes {
		_, err := tx.Exec(`
			INSERT INTO participant_byes (tournament_id, participant_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, t.ID, pid)
		if err != nil {
			return err
		}
	}

	// save matches
	for round, matches := range t.Matches {
		for _, m := range matches {
			var opinionP1, opinionP2, result *string
			if m.OpinionP1 != nil {
				v := string(*m.OpinionP1)
				opinionP1 = &v
			}
			if m.OpinionP2 != nil {
				v := string(*m.OpinionP2)
				opinionP2 = &v
			}
			if m.Result != nil {
				v := string(*m.Result)
				result = &v
			}

			_, err := tx.Exec(`
				INSERT INTO matches (id, tournament_id, round_number, p1_id, p2_id, state, opinion_p1, opinion_p2, result, scheduled_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
				ON CONFLICT (id) DO UPDATE
				    SET state = EXCLUDED.state,
				        opinion_p1 = EXCLUDED.opinion_p1,
				        opinion_p2 = EXCLUDED.opinion_p2,
				        result = EXCLUDED.result,
				        scheduled_at = EXCLUDED.scheduled_at
			`, m.ID, t.ID, round, m.P1, m.P2, m.State, opinionP1, opinionP2, result, m.ScheduledAt)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (s *PostgresStore) GetTournament(id domain.TournamentID) (*domain.Tournament, error) {
	row := s.db.QueryRow(`
		SELECT id, owner_id, title, system, current_round, last_round, start_time
		FROM tournaments WHERE id = $1
	`, id)

	t := &domain.Tournament{
		Matches:   make(map[domain.Round][]*domain.Match),
		Opponents: make(map[domain.ParticipantID]map[domain.ParticipantID]bool),
		Byes:      make(map[domain.ParticipantID]bool),
	}

	if err := row.Scan(&t.ID, &t.OwnerID, &t.Title, &t.System, &t.CurrentRound, &t.LastRound, &t.StartTime); err != nil {
		return nil, err
	}

	// load participants
	rows, err := s.db.Query(`
		SELECT id, kind, name, telegram_tag, eliminated, score, joined_at
		FROM participants WHERE tournament_id = $1 ORDER BY id
	`, t.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		p := &domain.Participant{}
		if err := rows.Scan(&p.ID, &p.Kind, &p.Name, &p.TelegramTag, &p.Eliminated, &p.Score, &p.JoinedAt); err != nil {
			return nil, err
		}
		p.TournamentID = t.ID

		// load roster
		members, err := s.db.Query(`SELECT telegram_user_id FROM participant_members WHERE participant_id = $1 AND tournament_id = $2`, p.ID, t.ID)
		if err != nil {
			return nil, err
		}
		for members.Next() {
			var uid domain.TelegramUserID
			if err := members.Scan(&uid); err != nil {
				return nil, err
			}
			p.Roster = append(p.Roster, uid)
		}
		members.Close()

		t.Participants = append(t.Participants, p)
	}

	// load matches
	matches, err := s.db.Query(`
		SELECT id, round_number, p1_id, p2_id, state, opinion_p1, opinion_p2, result, scheduled_at
		FROM matches WHERE tournament_id = $1
		ORDER BY round_number, id
	`, t.ID)
	if err != nil {
		return nil, err
	}
	defer matches.Close()

	for matches.Next() {
		m := &domain.Match{TournamentID: t.ID}
		var opinionP1, opinionP2, result sql.NullString
		if err := matches.Scan(&m.ID, &m.Round, &m.P1, &m.P2, &m.State, &opinionP1, &opinionP2, &result, &m.ScheduledAt); err != nil {
			return nil, err
		}

		if opinionP1.Valid {
			val := domain.ResultType(opinionP1.String)
			m.OpinionP1 = &val
		}
		if opinionP2.Valid {
			val := domain.ResultType(opinionP2.String)
			m.OpinionP2 = &val
		}
		if result.Valid {
			val := domain.ResultType(result.String)
			m.Result = &val
		}
		t.Matches[m.Round] = append(t.Matches[m.Round], m)
	}

	// load opponents history
	oppRows, err := s.db.Query(`
		SELECT participant1_id, participant2_id
		FROM participant_opponents WHERE tournament_id = $1
	`, t.ID)
	if err != nil {
		return nil, err
	}
	defer oppRows.Close()

	for oppRows.Next() {
		var p1, p2 domain.ParticipantID
		if err := oppRows.Scan(&p1, &p2); err != nil {
			return nil, err
		}
		if t.Opponents[p1] == nil {
			t.Opponents[p1] = make(map[domain.ParticipantID]bool)
		}
		if t.Opponents[p2] == nil {
			t.Opponents[p2] = make(map[domain.ParticipantID]bool)
		}
		t.Opponents[p1][p2] = true
		t.Opponents[p2][p1] = true
	}

	// load byes
	byeRows, err := s.db.Query(`
		SELECT participant_id FROM participant_byes WHERE tournament_id = $1
	`, t.ID)
	if err != nil {
		return nil, err
	}
	defer byeRows.Close()

	for byeRows.Next() {
		var p domain.ParticipantID
		if err := byeRows.Scan(&p); err != nil {
			return nil, err
		}
		t.Byes[p] = true
	}

	return t, nil
}

func (s *PostgresStore) GetTournaments() ([]*domain.Tournament, error) {
	rows, err := s.db.Query(`
		SELECT DISTINCT id
		FROM tournaments
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []domain.TournamentID
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, domain.TournamentID(id))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var tournaments []*domain.Tournament
	for _, tid := range ids {
		t, err := s.GetTournament(tid)
		if err != nil {
			return nil, err
		}
		tournaments = append(tournaments, t)
	}

	return tournaments, nil
}

func (s *PostgresStore) GetUserTournaments(userID domain.TelegramUserID) (map[domain.TournamentID]string, error) {
	rows, err := s.db.Query(`
		SELECT DISTINCT t.id, t.title
		FROM tournaments t
		LEFT JOIN participants p ON t.id = p.tournament_id
		LEFT JOIN participant_members pm ON p.id = pm.participant_id AND p.tournament_id = pm.tournament_id
		WHERE pm.telegram_user_id = $1 OR t.owner_id = $1
		ORDER BY t.id
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tournaments := make(map[domain.TournamentID]string)
	for rows.Next() {
		var id int64
		var title string
		if err := rows.Scan(&id, &title); err != nil {
			return nil, err
		}
		tournaments[domain.TournamentID(id)] = title
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tournaments, nil
}

func (s *PostgresStore) AddParticipant(p *domain.Participant) error {
	_, err := s.db.Exec(`
        INSERT INTO participants (id, tournament_id, kind, name, telegram_tag, eliminated, score, joined_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `, p.ID, p.TournamentID, p.Kind, p.Name, p.TelegramTag, p.Eliminated, p.Score, p.JoinedAt)
	if err != nil {
		return err
	}

	for _, tgid := range p.Roster {
		_, err = s.db.Exec(`INSERT INTO participant_members (tournament_id, participant_id, telegram_user_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`, p.TournamentID, p.ID, tgid)
		if err != nil {
			return err
		}
	}

	return nil
}
