-- +goose Up
ALTER TABLE users
ADD is_chirpy_red BOOLEAN DEFAULT false;

UPDATE users
SET is_chirpy_red = false;

ALTER TABLE users
ALTER COLUMN is_chirpy_red 
SET NOT NULL;

-- +goose Down
ALTER TABLE users
DROP COLUMN is_chirpy_red;