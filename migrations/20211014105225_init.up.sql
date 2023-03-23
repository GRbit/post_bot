BEGIN;

CREATE TABLE addresses (
    id          SERIAL PRIMARY KEY,
    telegram    UNIQUE TEXT,
    instagram   TEXT,
    person_name TEXT,
    address     TEXT,
    wishes      TEXT,
    phone       TEXT,
    email       TEXT,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
CREATE INDEX bookings_phone_idx ON addresses USING btree (phone);
CREATE INDEX bookings_email_idx ON addresses USING btree (email);
CREATE INDEX bookings_telegram_idx ON addresses USING btree (telegram);

CREATE TABLE users (
    id                 SERIAL PRIMARY KEY,
    chat_id            BIGINT,
    requested          BOOLEAN,
    received_addresses TEXT,
    search_previous    BOOLEAN,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone
);
CREATE UNIQUE INDEX users_chat_id_idx ON users USING btree (chat_id);

COMMIT;
