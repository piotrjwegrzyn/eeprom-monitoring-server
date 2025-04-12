-- name: CreateDevice :exec
INSERT INTO devices (hostname, ip, login, passwd, keyfile, connected)
VALUES (sqlc.arg(hostname), sqlc.arg(ip), sqlc.arg(login), sqlc.arg(passwd), sqlc.arg(keyfile), sqlc.arg(connected));

-- name: Device :one
SELECT * FROM devices
WHERE devices.id = sqlc.arg(id);

-- name: Devices :many
SELECT * FROM devices;

-- name: UpdateDevice :exec
UPDATE devices
SET hostname    = sqlc.arg(hostname),
    ip          = sqlc.arg(ip),
    login       = sqlc.arg(login),
    passwd      = sqlc.arg(passwd),
    keyfile     = sqlc.arg(keyfile),
    last_status = sqlc.arg(last_status),
    connected   = sqlc.arg(connected)
WHERE devices.id = sqlc.arg(id);

-- name: UpdateDeviceStatus :exec
UPDATE devices
SET last_status = sqlc.arg(last_status),
    connected   = sqlc.arg(connected)
WHERE devices.id = sqlc.arg(id);

-- name: DeleteDevice :exec
DELETE FROM devices
WHERE devices.id = sqlc.arg(id);