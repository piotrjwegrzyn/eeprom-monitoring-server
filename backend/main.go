package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"

	_ "github.com/go-sql-driver/mysql"

	"pi-wegrzyn/backend/cmds"
	"pi-wegrzyn/storage"
	"pi-wegrzyn/utils"
)

type ctxKey string

const appNameAttr ctxKey = "app"

func main() {
	appCtx := context.WithValue(context.Background(), appNameAttr, "backend")

	var configPath = flag.String("c", "config.yaml", "Path to config file (YAML file)")
	flag.Parse()

	if err := utils.StatPaths([]string{*configPath}); err != nil {
		slog.ErrorContext(appCtx, "cannot use provided path", slog.Any("error", err))
		os.Exit(1)
	}

	var cfg utils.Config
	if err := utils.ReadConfig(*configPath, &cfg); err != nil {
		slog.ErrorContext(appCtx, "cannot read configuration", slog.Any("error", err))
		os.Exit(1)
	}

	conn, closeConn, err := connectToDatabase(cfg.Database)
	if err != nil {
		slog.Error("cannot connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer closeConn()

	server := cmds.NewServer(&cfg, storage.New(conn))
	if err := server.Loop(appCtx); err != nil {
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
