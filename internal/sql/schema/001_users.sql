-- +goose Up
CREATE TABLE users (
	user_id UUID PRIMARY KEY NOT NULL,
	username TEXT NOT NULL,
	hashed_pw TEXT NOT NULL,
	admin BOOLEAN DEFAULT false NOT NULL,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE users;
