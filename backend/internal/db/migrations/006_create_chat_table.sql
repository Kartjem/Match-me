CREATE TABLE IF NOT EXISTS chats (
  id SERIAL PRIMARY KEY,
  sender_id INT REFERENCES users(id) ON DELETE CASCADE,
  receiver_id INT REFERENCES users(id) ON DELETE CASCADE,
  message TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  read BOOLEAN DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_chats_receiver_sender ON chats (receiver_id, sender_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_chats_read ON chats (receiver_id, read);
CREATE INDEX IF NOT EXISTS idx_chats_created_at ON chats (created_at);
