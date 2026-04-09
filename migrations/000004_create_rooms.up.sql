CREATE TABLE rooms
(
    id                  BIGSERIAL PRIMARY KEY,
    name                VARCHAR(255)   NOT NULL,
    slug                VARCHAR(255)   NOT NULL,
    room_type           room_type      NOT NULL,

    description         TEXT,
    short_description   VARCHAR(500),

    max_guests          INTEGER        NOT NULL DEFAULT 2 CHECK (max_guests > 0),
    num_bedrooms        INTEGER        NOT NULL DEFAULT 1 CHECK (num_bedrooms >= 0),
    num_bathrooms       INTEGER        NOT NULL DEFAULT 1 CHECK (num_bathrooms >= 0),
    num_beds            INTEGER        NOT NULL DEFAULT 1 CHECK (num_beds > 0),

    area                DECIMAL(6, 1) CHECK (area >= 0),

    address             TEXT,
    district            VARCHAR(100),
    city                VARCHAR(100)   NOT NULL DEFAULT 'Hà Nội',

    latitude            DECIMAL(10, 7),
    longitude           DECIMAL(10, 7),

    base_price          DECIMAL(12, 0) NOT NULL CHECK (base_price >= 0),
    cleaning_fee        DECIMAL(12, 0) NOT NULL DEFAULT 0 CHECK (cleaning_fee >= 0),

    min_nights          INTEGER        NOT NULL DEFAULT 1 CHECK (min_nights > 0),
    max_nights          INTEGER        NOT NULL DEFAULT 30,

    checkin_time        VARCHAR(10)    NOT NULL DEFAULT '14:00',
    checkout_time       VARCHAR(10)    NOT NULL DEFAULT '12:00',

    house_rules         TEXT,
    cancellation_policy TEXT,

    status              room_status    NOT NULL DEFAULT 'active',
    sort_order          INTEGER        NOT NULL DEFAULT 0,

    created_at          TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ,

    CONSTRAINT chk_rooms_max_nights CHECK (max_nights >= min_nights)
);

CREATE UNIQUE INDEX idx_rooms_slug ON rooms (slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_rooms_room_type ON rooms (room_type);
CREATE INDEX idx_rooms_status ON rooms (status);
CREATE INDEX idx_rooms_district ON rooms (district);
CREATE INDEX idx_rooms_city ON rooms (city);
CREATE INDEX idx_rooms_base_price ON rooms (base_price);
CREATE INDEX idx_rooms_sort_order ON rooms (sort_order);