-- +goose Up
CREATE TABLE apps (
    instance       TEXT NOT NULL,
    app_name       TEXT NOT NULL,
    app_id         TEXT NOT NULL,
    client_id      TEXT NOT NULL,
    client_secret  TEXT NOT NULL,
    access_token   TEXT,

    UNIQUE (instance, app_name)
);

-- +goose Down
DROP TABLE apps;
