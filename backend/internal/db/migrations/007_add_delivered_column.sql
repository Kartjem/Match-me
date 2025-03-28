ALTER TABLE chats ADD COLUMN IF NOT EXISTS delivered BOOLEAN DEFAULT false;
CREATE INDEX IF NOT EXISTS idx_chats_delivered ON chats (delivered);
