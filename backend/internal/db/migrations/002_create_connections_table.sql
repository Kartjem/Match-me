CREATE TABLE IF NOT EXISTS connections (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES users(id) ON DELETE CASCADE,
  connected_user_id INT REFERENCES users(id) ON DELETE CASCADE,
  status VARCHAR(10) NOT NULL DEFAULT 'pending',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_connections_unique ON connections (user_id, connected_user_id);
CREATE INDEX IF NOT EXISTS idx_connections_status ON connections (user_id, connected_user_id, status);
CREATE INDEX IF NOT EXISTS idx_connections_created_at ON connections (created_at);
