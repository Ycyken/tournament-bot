CREATE TABLE tournaments (
                             id BIGSERIAL PRIMARY KEY,
                             owner_id BIGINT NOT NULL,
                             title VARCHAR(255) NOT NULL,
                             system VARCHAR(50) NOT NULL, -- 'single_elimination', 'swiss'
                             current_round INT DEFAULT 0,
                             last_round INT DEFAULT 0,
                             start_time TIMESTAMP DEFAULT NOW(),
                             created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE participants (
                              id BIGINT NOT NULL,
                              tournament_id BIGINT NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
                              PRIMARY KEY (tournament_id, id),
                              kind VARCHAR(10) NOT NULL, -- 'user', 'team'
                              name VARCHAR(255) NOT NULL,
                              telegram_tag VARCHAR(255),
                              eliminated BOOLEAN DEFAULT false,
                              score DECIMAL(4,1) DEFAULT 0.0,
                              joined_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE participant_members (
                                     tournament_id BIGINT NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
                                     participant_id BIGINT NOT NULL,
                                     telegram_user_id BIGINT NOT NULL,
                                     PRIMARY KEY (tournament_id, participant_id, telegram_user_id),
                                     FOREIGN KEY (tournament_id, participant_id) REFERENCES participants(tournament_id, id) ON DELETE CASCADE
);

CREATE TABLE participant_opponents (
                                       tournament_id BIGINT NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
                                       participant1_id BIGINT NOT NULL,
                                       participant2_id BIGINT NOT NULL,
                                       FOREIGN KEY (tournament_id, participant1_id) REFERENCES participants(tournament_id, id),
                                       FOREIGN KEY (tournament_id, participant2_id) REFERENCES participants(tournament_id, id),
                                       PRIMARY KEY (tournament_id, participant1_id, participant2_id)
);

CREATE TABLE participant_byes (
                                  tournament_id BIGINT NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
                                  participant_id BIGINT NOT NULL,
                                  PRIMARY KEY (tournament_id, participant_id),
                                  FOREIGN KEY (tournament_id, participant_id) REFERENCES participants(tournament_id, id) ON DELETE CASCADE
);

CREATE TABLE matches (
                         id BIGINT,
                         tournament_id BIGINT NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
                         PRIMARY KEY (id, tournament_id),
                         round_number INT NOT NULL,
                         p1_id BIGINT NOT NULL,
                         p2_id BIGINT NOT NULL,
                         FOREIGN KEY (tournament_id, p1_id) REFERENCES participants(tournament_id, id),
                         FOREIGN KEY (tournament_id, p2_id) REFERENCES participants(tournament_id, id),
                         state VARCHAR(50) DEFAULT 'scheduled', -- 'scheduled', 'completed', 'conflicted'
                         opinion_p1 VARCHAR(10), -- 'p1', 'p2', 'draw'
                         opinion_p2 VARCHAR(10), -- 'p1', 'p2', 'draw'
                         result VARCHAR(10),     -- 'p1', 'p2', 'draw'
                         scheduled_at TIMESTAMP,
                         created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE applications (
                              tournament_id BIGINT NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
                              telegram_user_id BIGINT NOT NULL,
                              name VARCHAR(255) NOT NULL,
                              telegram_tag VARCHAR(255),
                              text TEXT,
                              created_at TIMESTAMP DEFAULT NOW(),
                              PRIMARY KEY (tournament_id, telegram_user_id)
);

CREATE INDEX idx_tournaments_owner ON tournaments(owner_id);
CREATE INDEX idx_participants_tournament ON participants(tournament_id);
CREATE INDEX idx_participant_members_participant
    ON participant_members(tournament_id, participant_id);
CREATE INDEX idx_participant_members_user
    ON participant_members(telegram_user_id);
CREATE INDEX idx_matches_tournament
    ON matches(tournament_id);
CREATE INDEX idx_matches_round
    ON matches(tournament_id, round_number);
CREATE INDEX idx_applications_user
    ON applications(telegram_user_id);
