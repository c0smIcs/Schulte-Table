CREATE TABLE record
(
    ID        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    SessionID VARCHAR(250)     NOT NULL,
    TimeTaken DOUBLE PRECISION NOT NULL,
    CreatedAt TIMESTAMP        NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_record_session_id ON record (SessionID);