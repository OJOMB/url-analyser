package config

// Config holds all application configuration values
type Config struct {
	Env    string
	IP     string
	Port   uint
	App    string
	Static string
}

// ConfigMap holds the configuration data for each given environment
var ConfigMap = map[string]Config{
	"dev": Config{
		Env:    "dev",
		IP:     "0.0.0.0",
		Port:   8080,
		App:    "UrlAnalyser",
		Static: "./public",
	},
}
