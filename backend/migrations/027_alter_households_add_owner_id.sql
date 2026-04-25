ALTER TABLE households ADD COLUMN owner_id TEXT REFERENCES users(id);
