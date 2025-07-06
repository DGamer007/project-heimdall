-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS core.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(254) NOT NULL UNIQUE,
    username VARCHAR(254) NOT NULL UNIQUE,
    password VARCHAR(254) NOT NULL CHECK (char_length(password) >= 8),
    first_name VARCHAR(254),
    last_name VARCHAR(254),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER trigger_users_updated_at
BEFORE UPDATE ON core.users
FOR EACH ROW
EXECUTE FUNCTION modify_timestamp();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trigger_users_updated_at ON core.users;

DROP TABLE IF EXISTS core.users;
-- +goose StatementEnd
