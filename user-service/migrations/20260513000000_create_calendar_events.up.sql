CREATE TABLE calendar_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_uuid UUID NOT NULL,
    date DATE NOT NULL,
    time TIME NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,

    FOREIGN KEY (user_uuid)
    REFERENCES users(uuid)
    ON DELETE CASCADE
);

CREATE INDEX idx_calendar_events_user_date ON calendar_events (user_uuid, date);
