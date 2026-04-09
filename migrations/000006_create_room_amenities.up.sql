CREATE TABLE room_amenities
(
    room_id    BIGINT NOT NULL REFERENCES rooms (id) ON DELETE CASCADE,
    amenity_id BIGINT NOT NULL REFERENCES amenities (id) ON DELETE CASCADE,

    PRIMARY KEY (room_id, amenity_id)
);

CREATE INDEX idx_room_amenities_amenity_id ON room_amenities (amenity_id);