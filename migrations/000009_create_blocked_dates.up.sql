CREATE TABLE blocked_dates
(
    id         BIGSERIAL PRIMARY KEY,
    room_id    BIGINT       NOT NULL REFERENCES rooms (id) ON DELETE CASCADE,
    date       DATE         NOT NULL,
    source     block_source NOT NULL,

    source_ref VARCHAR(255),
    reason     VARCHAR(255),

    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT uq_blocked_dates_room_date_src UNIQUE (room_id, date, source)
);

CREATE INDEX idx_blocked_dates_source ON blocked_dates (source);
CREATE INDEX idx_blocked_dates_room_date ON blocked_dates (room_id, date);