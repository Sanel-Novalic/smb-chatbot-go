CREATE TABLE IF NOT EXISTS conversations (
    chat_id BIGINT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    state VARCHAR(50) NOT NULL,
    last_interaction_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS reviews (
    id UUID PRIMARY KEY,
    customer_id BIGINT NOT NULL,
    chat_id BIGINT NOT NULL,
    text TEXT NOT NULL,
    received_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS message_history (
    id SERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL,
    is_user_message BOOLEAN NOT NULL,
    text TEXT NOT NULL,
    "timestamp" TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (chat_id) REFERENCES conversations(chat_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_message_history_chat_id_timestamp ON message_history (chat_id, "timestamp");