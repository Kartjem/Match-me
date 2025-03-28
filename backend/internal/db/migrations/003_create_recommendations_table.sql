CREATE TABLE IF NOT EXISTS recommendations (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES users(id) ON DELETE CASCADE,
  recommended_user_id INT REFERENCES users(id) ON DELETE CASCADE,
  score DECIMAL(5, 2) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_recommendations_user_id ON recommendations (user_id);
CREATE INDEX IF NOT EXISTS idx_recommendations_score ON recommendations (score DESC);
