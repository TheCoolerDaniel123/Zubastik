CREATE TABLE tooth_notes (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    doctor_uuid  UUID NOT NULL,
    patient_uuid UUID NOT NULL,
    tooth_index  SMALLINT NOT NULL,
    note         TEXT NOT NULL,
    created_at   TIMESTAMP NOT NULL DEFAULT NOW(),

    FOREIGN KEY (patient_uuid) REFERENCES users(uuid) ON DELETE CASCADE
);

CREATE INDEX idx_tooth_notes_patient_tooth
    ON tooth_notes (patient_uuid, tooth_index);

CREATE TABLE recommendations (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    doctor_uuid  UUID NOT NULL,
    patient_uuid UUID NOT NULL,
    text         TEXT NOT NULL,
    created_at   TIMESTAMP NOT NULL DEFAULT NOW(),

    FOREIGN KEY (patient_uuid) REFERENCES users(uuid) ON DELETE CASCADE
);
