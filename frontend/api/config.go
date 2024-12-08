package api

type Config struct {
	User         string `envconfig:"ADMIN_USER"`
	Password     string `envconfig:"ADMIN_PASSWORD"`
	Port         string `envconfig:"APP_PORT"`
	TemplatesDir string `envconfig:"TEMPLATES_DIR" default:"/etc/ems/templates/"`
	StaticDir    string `envconfig:"STATIC_DIR default:"/etc/ems/static/"`
}
