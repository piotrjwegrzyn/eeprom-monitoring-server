-- +goose UP
-- +goose StatementBegin
CREATE TABLE devices
(
  id          INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  hostname    VARCHAR(100) NOT NULL,
  ip          VARCHAR(100) NOT NULL,
  login       VARCHAR(100) NOT NULL,
  passwd      VARCHAR(100) DEFAULT NULL,
  keyfile     BLOB DEFAULT NULL,
  connected   DATETIME NOT NULL,
  last_status INT NOT NULL DEFAULT -1
) COLLATE = utf8mb4_unicode_ci CHARSET = utf8mb4 COMMENT 'Network devices set up for monitoring';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE devices;
-- +goose StatementEnd
