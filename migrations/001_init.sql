CREATE TABLE users (
                       id BIGSERIAL PRIMARY KEY,
                       telegram_id BIGINT UNIQUE NOT NULL,
                       username VARCHAR(255),
                       first_name VARCHAR(255),
                       created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE tournaments (
                             id BIGSERIAL PRIMARY KEY,
                             owner_id BIGINT NOT NULL REFERENCES users(telegram_id),
                             title VARCHAR(255) NOT NULL,
                             system VARCHAR(50) NOT NULL, -- 'single_elimination' или 'swiss'
                             current_round INT DEFAULT 0,
                             last_round INT DEFAULT 0,
                             start_time TIMESTAMP DEFAULT NOW(),
                             created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE participants (
                              id BIGSERIAL PRIMARY KEY,
                              tournament_id BIGINT NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
                              kind VARCHAR(10) NOT NULL, -- 'user', 'team'
                              name VARCHAR(255) NOT NULL,
                              eliminated BOOLEAN DEFAULT false,
                              score DECIMAL(4,1) DEFAULT 0.0,
                              joined_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE participant_members (
                                     id BIGSERIAL PRIMARY KEY,
                                     participant_id BIGINT NOT NULL REFERENCES participants(id) ON DELETE CASCADE,
                                     telegram_user_id BIGINT NOT NULL REFERENCES users(telegram_id),
                                     UNIQUE(participant_id, telegram_user_id)
);

CREATE TABLE participant_opponents (
                                       id BIGSERIAL PRIMARY KEY,
                                       tournament_id BIGINT NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
                                       participant1_id BIGINT NOT NULL REFERENCES participants(id),
                                       participant2_id BIGINT NOT NULL REFERENCES participants(id),
                                       UNIQUE(tournament_id, participant1_id, participant2_id)
);

CREATE TABLE participant_byes (
                                  id BIGSERIAL PRIMARY KEY,
                                  tournament_id BIGINT NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
                                  participant_id BIGINT NOT NULL REFERENCES participants(id),
                                  round_number INT NOT NULL,
                                  UNIQUE(tournament_id, participant_id, round_number)
);



CREATE TABLE matches (
                         id BIGSERIAL PRIMARY KEY,
                         tournament_id BIGINT NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
                         round_number INT NOT NULL,
                         p1_id BIGINT NOT NULL REFERENCES participants(id),
                         p2_id BIGINT NOT NULL REFERENCES participants(id),
                         state VARCHAR(50) DEFAULT 'scheduled', -- 'scheduled', 'completed', 'conflicted'
                         opinion_p1 VARCHAR(10), -- 'p1', 'p2', 'draw'
                         opinion_p2 VARCHAR(10), -- 'p1', 'p2', 'draw'
                         result VARCHAR(10),     -- 'p1', 'p2', 'draw'
                         scheduled_at TIMESTAMP,
                         created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE applications (
                              id BIGSERIAL PRIMARY KEY,
                              tournament_id BIGINT NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
                              telegram_user_id BIGINT NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE,
                              created_at TIMESTAMP DEFAULT NOW(),
                              UNIQUE (tournament_id, telegram_user_id)
);

CREATE INDEX idx_users_telegram_id ON users(telegram_id);
CREATE INDEX idx_tournaments_owner ON tournaments(owner_id);
CREATE INDEX idx_participants_tournament ON participants(tournament_id);
CREATE INDEX idx_matches_tournament ON matches(tournament_id);
CREATE INDEX idx_matches_round ON matches(tournament_id, round_number);
CREATE INDEX idx_applications_tournament ON applications(tournament_id);
CREATE INDEX idx_applications_user ON applications(telegram_user_id);