CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS players (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id),
  name TEXT NOT NULL UNIQUE,
  best_score INTEGER NOT NULL DEFAULT 0,
  color INTEGER NOT NULL
);

-- Index for faster leaderboard queries
CREATE INDEX IF NOT EXISTS idx_players_best_score ON players(best_score DESC);

