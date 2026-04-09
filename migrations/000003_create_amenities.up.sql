CREATE TABLE amenities
(
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(100) NOT NULL,
    icon       VARCHAR(255),
    category   VARCHAR(50),

    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_amenities_name ON amenities (name) WHERE deleted_at IS NULL;
CREATE INDEX idx_amenities_category ON amenities (category);