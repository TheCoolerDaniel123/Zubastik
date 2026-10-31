CREATE TABLE IF NOT EXISTS conversations (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    doctor_uuid  UUID NOT NULL REFERENCES users(uuid) ON DELETE CASCADE,
    patient_uuid UUID NOT NULL REFERENCES users(uuid) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (doctor_uuid, patient_uuid)
);

CREATE INDEX IF NOT EXISTS idx_conversations_doctor  ON conversations (doctor_uuid);
CREATE INDEX IF NOT EXISTS idx_conversations_patient ON conversations (patient_uuid);

CREATE TABLE IF NOT EXISTS direct_messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_uuid     UUID NOT NULL REFERENCES users(uuid) ON DELETE CASCADE,
    content         TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_direct_messages_conv ON direct_messages (conversation_id, created_at);
