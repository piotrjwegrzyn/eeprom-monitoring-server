package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/kelseyhightower/envconfig"

	"pi-wegrzyn/frontend/api"
	"pi-wegrzyn/storage"
)

type ctxKey string

const appNameAttr ctxKey = "app"

func main() {
	appCtx := context.WithValue(context.Background(), appNameAttr, "frontend")

	var cfg api.Config
	if err := envconfig.Process("", &cfg); err != nil {
		slog.ErrorContext(appCtx, "cannot read configuration", slog.Any("error", err))
		os.Exit(1)
	}

	var dbCfg storage.Config
	if err := envconfig.Process("", &dbCfg); err != nil {
		slog.ErrorContext(appCtx, "cannot read database configuration", slog.Any("error", err))
		os.Exit(1)
	}

	conn, closeConn, err := connectToDatabase(dbCfg)
	if err != nil {
		slog.Error("cannot connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer closeConn()

	cookies := make(map[string]api.Cookie)
	handler, err := api.NewServerAPI(cfg, storage.New(conn), &cookies)
	if err != nil {
		slog.ErrorContext(appCtx, "cannot prepare server", slog.Any("error", err))
		os.Exit(1)
	}

	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		slog.ErrorContext(appCtx, "server failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func connectToDatabase(config storage.Config) (conn *sql.DB, closeConn func(), err error) {
	conn, err = sql.Open("mysql", config.String())
	if err != nil {
		return conn, closeConn, err
	}

	closeConn = func() {
		if err = conn.Close(); err != nil {
			slog.Error("can't close database connection: " + err.Error())
		}
	}

	return conn, closeConn, nil
}
