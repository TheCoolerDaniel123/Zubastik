
CREATE TABLE users (
    uuid UUID PRIMARY KEY,
    phone VARCHAR(255) NOT NULL UNIQUE,
    role SMALLINT NOT NULL,

    name TEXT NULL,
    age INT NULL,
    gender TEXT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- INDEX users table

CREATE TABLE teeth (
    user_uuid UUID NOT NULL,
    tooth_index SMALLINT NOT NULL,
    tooth_type TEXT NOT NULL,
    tooth_state TEXT NOT NULL DEFAULT 'Healthy',

    PRIMARY KEY (user_uuid, tooth_index),

    FOREIGN KEY (user_uuid)
    REFERENCES users(uuid)
    ON DELETE CASCADE
);

-- INDEX teeth table

CREATE TABLE brushing_times (
    user_uuid UUID NOT NULL,
    time TIME NOT NULL,

    PRIMARY KEY (user_uuid, time),

    FOREIGN KEY (user_uuid)
    REFERENCES users(uuid)
    ON DELETE CASCADE
);


CREATE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();