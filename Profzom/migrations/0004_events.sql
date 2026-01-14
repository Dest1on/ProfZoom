-- +goose Up
CREATE TABLE analytics_events (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    user_id UUID,
    payload JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX idx_analytics_events_name ON analytics_events(name);
CREATE INDEX idx_analytics_events_user ON analytics_events(user_id);

-- +goose Down
DROP TABLE analytics_events;
