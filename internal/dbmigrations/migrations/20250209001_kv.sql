-- +goose Up
CREATE TABLE kv (
      instance       TEXT NOT NULL,
      app_name       TEXT NOT NULL,
      key            TEXT NOT NULL,
      value          TEXT NOT NULL,

      UNIQUE (instance, app_name, key)
);

-- +goose Down
DROP TABLE kv;
