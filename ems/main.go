package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/kelseyhightower/envconfig"

	"pi-wegrzyn/ems/api"
	"pi-wegrzyn/ems/cmds"
	"pi-wegrzyn/ems/cookies"
	"pi-wegrzyn/ems/influx"
	"pi-wegrzyn/ems/storage"
	"pi-wegrzyn/ems/templates"
)

type ctxKey string

const appNameAttr ctxKey = "app"

type config struct {
	Port     string `envconfig:"APP_PORT" default:"8080"`
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`

	apiConfig    api.Config
	dbConfig     storage.Config
	influxConfig influx.Config
	loopConfig   cmds.Config
	assetsConfig assetsConfig
}

type assetsConfig struct {
	TemplatesDir string `envconfig:"TEMPLATES_DIR" default:"/ems/templates/"`
	Style        string `envconfig:"STYLE_PATH" default:"/ems/static/style.css"`
	Favicon      string `envconfig:"FAVICON_PATH" default:"/ems/static/favicon.ico"`
}

func main() {
	appCtx := context.WithValue(context.Background(), appNameAttr, "ems")

	var config config
	if err := envconfig.Process("", &config); err != nil {
		slog.ErrorContext(appCtx, "cannot read configuration", slog.Any("error", err))
		os.Exit(1)
	}
	if err := envconfig.Process("API", &config.apiConfig); err != nil {
		slog.ErrorContext(appCtx, "cannot read configuration", slog.Any("error", err))
		os.Exit(1)
	}
	if err := envconfig.Process("DB", &config.dbConfig); err != nil {
		slog.ErrorContext(appCtx, "cannot read database configuration", slog.Any("error", err))
		os.Exit(1)
	}
	if err := envconfig.Process("ASSETS", &config.assetsConfig); err != nil {
		slog.ErrorContext(appCtx, "cannot read assets configuration", slog.Any("error", err))
		os.Exit(1)
	}
	if err := envconfig.Process("INFLUX", &config.influxConfig); err != nil {
		slog.ErrorContext(appCtx, "cannot read influx configuration", slog.Any("error", err))
		os.Exit(1)
	}

	tmplExecutor, err := templates.NewExecutor(config.assetsConfig.TemplatesDir)
	if err != nil {
		slog.ErrorContext(appCtx, "cannot initialize templates", slog.Any("error", err))
		os.Exit(1)
	}

	css, err := os.ReadFile(config.assetsConfig.Style)
	if err != nil {
		slog.ErrorContext(appCtx, "cannot read css file", slog.Any("error", err))
		os.Exit(1)
	}

	favicon, err := os.ReadFile(config.assetsConfig.Favicon)
	if err != nil {
		slog.ErrorContext(appCtx, "cannot read favicon file", slog.Any("error", err))
		os.Exit(1)
	}

	conn, closeConn, err := connectToDatabase(config.dbConfig)
	if err != nil {
		slog.Error("cannot connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer closeConn()

	influxClient, err := connectToInfluxDB(config.influxConfig)
	if err != nil {
		slog.Error("cannot connect to influxdb", slog.Any("error", err))
		os.Exit(1)
	}
	defer influxClient.Close()

	cookieStore := cookies.NewStore(15 * time.Minute)

	handler := api.NewServerAPI(
		config.apiConfig,
		storage.New(conn),
		cookieStore,
		tmplExecutor,
		&api.StaticFiles{
			CSS:     css,
			Favicon: favicon,
		},
	)
	server := cmds.NewServer(config.loopConfig, storage.New(conn), influx.New(config.influxConfig, influxClient))

	shutdownFunc := func(exitCode int) {
		appCtx.Done()
		os.Exit(exitCode)
	}
	defer shutdownFunc(0)

	go func() {
		if err := server.Loop(appCtx); err != nil {
			slog.ErrorContext(appCtx, "backend failed", slog.Any("error", err))
			shutdownFunc(1)
		}
	}()

	if err := http.ListenAndServe(":"+config.Port, handler); err != nil {
		slog.ErrorContext(appCtx, "http server failed", slog.Any("error", err))
		shutdownFunc(1)
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

func connectToInfluxDB(cfg influx.Config) (influxdb2.Client, error) {
	client := influxdb2.NewClient(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port), cfg.Token)
	_, err := client.Health(context.Background())
	return client, err
}
