package config

import (
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

// Version is set at build time via ldflags: -X dotkey-cli/internal/config.Version=<tag>
var Version = "dev"

type Config struct {
	APIURL             string `yaml:"api_url"`
	WebURL             string `yaml:"web_url"`
	Token              string `yaml:"token"`
	CurrentProjectID   string `yaml:"current_project_id"`
	CurrentProjectName string `yaml:"current_project_name"`
	CurrentEnvID       string `yaml:"current_env_id"`
	CurrentEnvName     string `yaml:"current_env_name"`
}

// LocalFile is .dotkey in the project directory — safe to commit, no secrets.
type LocalFile struct {
	Project     string `yaml:"project"`
	Environment string `yaml:"environment"`
}

func Dir() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "dotkey")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".dotkey")
}

func path() string { return filepath.Join(Dir(), "config.yaml") }

func Load() *Config {
	cfg := &Config{
		APIURL: "http://localhost:8080",
		WebURL: "http://localhost:3000",
	}

	data, err := os.ReadFile(path())
	if err == nil {
		yaml.Unmarshal(data, cfg) //nolint:errcheck
	}

	if cfg.WebURL == "" {
		cfg.WebURL = "http://localhost:3000"
	}

	// environment variable overrides — useful in CI
	if t := os.Getenv("DOTKEY_TOKEN"); t != "" {
		cfg.Token = t
	}
	if u := os.Getenv("DOTKEY_API_URL"); u != "" {
		cfg.APIURL = u
	}
	if u := os.Getenv("DOTKEY_WEB_URL"); u != "" {
		cfg.WebURL = u
	}

	return cfg
}

func Save(cfg *Config) error {
	if err := os.MkdirAll(Dir(), 0700); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path(), data, 0600)
}

func LoadLocalFile() *LocalFile {
	data, err := os.ReadFile(".dotkey")
	if err != nil {
		return nil
	}
	var lf LocalFile
	yaml.Unmarshal(data, &lf) //nolint:errcheck
	return &lf
}

func SaveLocalFile(lf *LocalFile) error {
	data, err := yaml.Marshal(lf)
	if err != nil {
		return err
	}
	return os.WriteFile(".dotkey", data, 0644)
}
