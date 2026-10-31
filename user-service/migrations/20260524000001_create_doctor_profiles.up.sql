CREATE TABLE IF NOT EXISTS doctor_profiles (
    user_uuid        UUID PRIMARY KEY REFERENCES users(uuid) ON DELETE CASCADE,
    specialization   TEXT NOT NULL DEFAULT '',
    experience_years INT  NOT NULL DEFAULT 0,
    rating           NUMERIC(3,1) NOT NULL DEFAULT 5.0,
    photo_url        TEXT NOT NULL DEFAULT ''
);
