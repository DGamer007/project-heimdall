-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS core.sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    access_token_jti UUID NOT NULL UNIQUE,
    refresh_token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES core.users(id) ON DELETE CASCADE
);

CREATE TRIGGER trigger_sessions_updated_at
BEFORE UPDATE ON core.sessions
FOR EACH ROW
EXECUTE FUNCTION modify_timestamp();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trigger_sessions_updated_at ON core.sessions;

DROP TABLE IF EXISTS core.sessions;
-- +goose StatementEnd
