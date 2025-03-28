CREATE TABLE IF NOT EXISTS dismissed_recommendations (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES users(id) ON DELETE CASCADE,
  dismissed_user_id INT REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_dismissed_user_pair
ON dismissed_recommendations (user_id, dismissed_user_id);
