CREATE TABLE articles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_uuid UUID NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    FOREIGN KEY (author_uuid) REFERENCES users(uuid) ON DELETE CASCADE
);

CREATE INDEX idx_articles_status ON articles (status);
CREATE INDEX idx_articles_author ON articles (author_uuid);
