-- +goose Up
BEGIN;

CREATE TABLE IF NOT EXISTS users (
     id          BIGSERIAL PRIMARY KEY,
     uuid        UUID         NOT NULL UNIQUE,
     name        TEXT         NOT NULL,
     created_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS products (
    id            BIGSERIAL PRIMARY KEY,
    title         TEXT          NOT NULL,
    description   TEXT,
    price         NUMERIC(10,2) NOT NULL,
    rating_votes  BIGINT        NOT NULL DEFAULT 0,
    rating        NUMERIC(3,2)  NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS reviews (
    id               BIGSERIAL PRIMARY KEY,
    user_id          BIGINT  NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
    product_id       BIGINT  NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    text             TEXT    NOT NULL,
    user_evaluation  SMALLINT NOT NULL,
    ai_evaluation    BOOLEAN  NULL,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT now()
    UNIQUE (user_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_reviews_user_id    ON reviews(user_id);
CREATE INDEX IF NOT EXISTS idx_reviews_product_id ON reviews(product_id);



COMMIT;

-- +goose Down
BEGIN;
DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS users;
COMMIT;