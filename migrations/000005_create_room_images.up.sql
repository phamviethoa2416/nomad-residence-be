CREATE TABLE room_images
(
    id         BIGSERIAL PRIMARY KEY,
    room_id    BIGINT       NOT NULL REFERENCES rooms (id) ON DELETE CASCADE,

    url        VARCHAR(500) NOT NULL,
    alt_text   VARCHAR(255),
    is_primary BOOLEAN      NOT NULL DEFAULT FALSE,
    sort_order INTEGER      NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_room_images_room_id ON room_images (room_id);