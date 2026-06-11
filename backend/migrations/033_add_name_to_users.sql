ALTER TABLE users ADD COLUMN name TEXT;

UPDATE users SET name = upper(substring(split_part(email, '@', 1) from 1 for 1)) || lower(substring(split_part(email, '@', 1) from 2));

ALTER TABLE users ALTER COLUMN name SET NOT NULL;
