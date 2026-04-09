-- Enable required extensions
CREATE
EXTENSION IF NOT EXISTS "pgcrypto";

-- Admin enums
CREATE TYPE admin_role AS ENUM ('admin', 'super_admin');

-- Room enums
CREATE TYPE room_type AS ENUM ('entire_place', 'private_room', 'shared_room', 'studio', 'apartment', 'villa', 'homestay');
CREATE TYPE room_status AS ENUM ('active', 'inactive', 'draft');

-- Booking enums
CREATE TYPE booking_status AS ENUM ('pending', 'confirmed', 'checked_in', 'checked_out', 'canceled', 'expired');
CREATE TYPE booking_source AS ENUM ('website', 'airbnb', 'booking', 'agoda', 'direct', 'other');
CREATE TYPE refund_status AS ENUM ('none', 'pending', 'partial', 'full', 'rejected');

-- Payment enums
CREATE TYPE payment_method AS ENUM ('vietqr', 'cash', 'bank_transfer', 'card', 'other');
CREATE TYPE payment_status AS ENUM ('pending', 'paid', 'failed', 'refunded', 'canceled');

-- BlockedDate enums
CREATE TYPE block_source AS ENUM ('manual', 'booking', 'ical', 'maintenance');

-- IcalLink enums
CREATE TYPE ical_sync_status AS ENUM ('idle', 'syncing', 'success', 'error');

-- PricingRule enums
CREATE TYPE pricing_rule_type AS ENUM ('date_range', 'day_of_week', 'last_minute', 'long_stay', 'seasonal', 'custom');
CREATE TYPE modifier_type AS ENUM ('fixed', 'percent');