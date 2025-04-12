package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/go-sql-driver/mysql"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"

	"pi-wegrzyn/backend/cmds"
	"pi-wegrzyn/backend/influx"
	"pi-wegrzyn/storage"
)

type ctxKey string

const appNameAttr ctxKey = "app"

func main() {
	appCtx := context.WithValue(context.Background(), appNameAttr, "backend")

	var configPath = flag.String("c", "config.yaml", "Path to config file (YAML file)")
	flag.Parse()

	var cfg cmds.Config
	if err := readConfig(*configPath, &cfg); err != nil {
		slog.ErrorContext(appCtx, "cannot read configuration", slog.Any("error", err))
		os.Exit(1)
	}

	var dbCfg storage.Config
	if err := envconfig.Process("", &dbCfg); err != nil {
		slog.ErrorContext(appCtx, "cannot read database configuration", slog.Any("error", err))
		os.Exit(1)
	}

	var influxCfg influx.Config
	if err := envconfig.Process("", &influxCfg); err != nil {
		slog.ErrorContext(appCtx, "cannot read influx configuration", slog.Any("error", err))
		os.Exit(1)
	}

	conn, closeConn, err := connectToDatabase(dbCfg)
	if err != nil {
		slog.Error("cannot connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer closeConn()

	client, err := connectToInfluxDB(influxCfg)
	if err != nil {
		slog.Error("cannot connect to influxdb", slog.Any("error", err))
		os.Exit(1)
	}
	defer client.Close()

	server := cmds.NewServer(cfg, storage.New(conn), influx.New(influxCfg, client))
	if err := server.Loop(appCtx); err != nil {
		slog.ErrorContext(appCtx, "server failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func connectToDatabase(cfg storage.Config) (conn *sql.DB, closeConn func(), err error) {
	conn, err = sql.Open("mysql", cfg.String())
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

func connectToInfluxDB(cfg influx.Config) (influxdb2.Client, error) {
	client := influxdb2.NewClient(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port), cfg.Token)
	_, err := client.Health(context.Background())
	return client, err
}

func readConfig(filename string, out any) error {
	cfg, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(cfg, out); err != nil {
		return err
	}

	return nil
}
