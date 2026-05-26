package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

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

// keyFile returns the path to the per-machine encryption key file.
// This file is created once on first login with mode 0600.
func keyFile() string { return filepath.Join(Dir(), "key") }

// loadOrCreateMachineKey reads the machine key from disk, creating it if absent.
// The key is 32 random bytes stored as a 64-char hex string.
func loadOrCreateMachineKey() (string, error) {
	kp := keyFile()
	data, err := os.ReadFile(kp)
	if err == nil {
		k := strings.TrimSpace(string(data))
		if len(k) == 64 {
			return k, nil
		}
	}
	// generate a new key
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate machine key: %w", err)
	}
	k := hex.EncodeToString(b)
	if err := os.MkdirAll(Dir(), 0700); err != nil {
		return "", err
	}
	if err := os.WriteFile(kp, []byte(k+"\n"), 0600); err != nil {
		return "", fmt.Errorf("write machine key: %w", err)
	}
	return k, nil
}

// ── Token encrypt/decrypt (AES-256-GCM, key = SHA-256(machineKey)) ───────────

const encPrefix = "encrypted:v1:"

func deriveKey(machineKey string) []byte {
	h := sha256.Sum256([]byte(machineKey))
	return h[:]
}

func encryptToken(machineKey, plaintext string) (string, error) {
	block, err := aes.NewCipher(deriveKey(machineKey))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ct := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	return fmt.Sprintf("%s%s:%s",
		encPrefix,
		base64.RawURLEncoding.EncodeToString(nonce),
		base64.RawURLEncoding.EncodeToString(ct),
	), nil
}

func decryptToken(machineKey, encoded string) (string, error) {
	if !strings.HasPrefix(encoded, encPrefix) {
		return encoded, nil // plaintext from older version — pass through
	}
	body := strings.TrimPrefix(encoded, encPrefix)
	parts := strings.SplitN(body, ":", 2)
	if len(parts) != 2 {
		return "", errors.New("malformed token in config")
	}
	nonce, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", err
	}
	ct, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(deriveKey(machineKey))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	pt, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", errors.New("failed to decrypt token — config may be corrupted")
	}
	return string(pt), nil
}

// ── Public API ────────────────────────────────────────────────────────────────

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

	// Decrypt token using machine key (transparent — old plaintext tokens still work)
	if cfg.Token != "" && strings.HasPrefix(cfg.Token, encPrefix) {
		mk, err := loadOrCreateMachineKey()
		if err == nil {
			if plain, err := decryptToken(mk, cfg.Token); err == nil {
				cfg.Token = plain
			}
		}
	}

	// Environment variable overrides — useful in CI
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

	// Encrypt the token before writing to disk
	toWrite := *cfg
	if toWrite.Token != "" && !strings.HasPrefix(toWrite.Token, encPrefix) {
		mk, err := loadOrCreateMachineKey()
		if err != nil {
			return fmt.Errorf("machine key: %w", err)
		}
		enc, err := encryptToken(mk, toWrite.Token)
		if err != nil {
			return fmt.Errorf("encrypt token: %w", err)
		}
		toWrite.Token = enc
	}

	data, err := yaml.Marshal(toWrite)
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
