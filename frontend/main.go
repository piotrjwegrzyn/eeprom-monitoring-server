package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/kelseyhightower/envconfig"

	"pi-wegrzyn/frontend/api"
	"pi-wegrzyn/frontend/cookies"
	"pi-wegrzyn/frontend/templates"
	"pi-wegrzyn/storage"
)

type ctxKey string

const appNameAttr ctxKey = "app"

type config struct {
	Port     string `envconfig:"APP_PORT" default:"8080"`
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`

	apiConfig    api.Config
	dbConfig     storage.Config
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

	if err := http.ListenAndServe(":"+config.Port, handler); err != nil {
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
