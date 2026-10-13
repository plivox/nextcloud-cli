package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/plivox/nextcloud-cli/internal/api"
	"github.com/zalando/go-keyring"
)

const (
	keyringService = "nextcloud-cli"
	keyringUser    = "default"
)

var ErrNotLoggedIn = errors.New(
	"not logged in: run `nextcloud login <server>` or set NEXTCLOUD_CLI_BASE_URL, NEXTCLOUD_CLI_USERNAME and NEXTCLOUD_CLI_PASSWORD",
)

type Credentials struct {
	Server      string `json:"server"`
	LoginName   string `json:"loginName"`
	AppPassword string `json:"appPassword"`
	UserID      string `json:"userId"`
}

func LoadCredentials() (Credentials, error) {
	baseURL := os.Getenv("NEXTCLOUD_CLI_BASE_URL")
	username := os.Getenv("NEXTCLOUD_CLI_USERNAME")
	password := os.Getenv("NEXTCLOUD_CLI_PASSWORD")

	if baseURL != "" || username != "" || password != "" {
		if baseURL == "" || username == "" || password == "" {
			return Credentials{}, fmt.Errorf(
				"NEXTCLOUD_CLI_BASE_URL, NEXTCLOUD_CLI_USERNAME and NEXTCLOUD_CLI_PASSWORD must all be set (via .env or the environment)",
			)
		}

		return Credentials{
			Server:      baseURL,
			LoginName:   username,
			AppPassword: password,
			UserID:      username,
		}, nil
	}

	return LoadStoredCredentials()
}

func NewClient() (*api.Client, error) {
	creds, err := LoadCredentials()
	if err != nil {
		return nil, err
	}

	return creds.Client(), nil
}

func (c Credentials) Client() *api.Client {
	return api.NewClient(c.Server, c.LoginName, c.AppPassword, c.UserID)
}

func CredentialsFromEnv() bool {
	return os.Getenv("NEXTCLOUD_CLI_BASE_URL") != "" ||
		os.Getenv("NEXTCLOUD_CLI_USERNAME") != "" ||
		os.Getenv("NEXTCLOUD_CLI_PASSWORD") != ""
}

func credentialsFile() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "nextcloud-cli", "credentials.json"), nil
}

func LoadStoredCredentials() (Credentials, error) {
	secret, err := keyring.Get(keyringService, keyringUser)
	if err == nil {
		return parseCredentials([]byte(secret), "the keyring")
	}

	path, err := credentialsFile()
	if err != nil {
		return Credentials{}, err
	}

	content, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Credentials{}, ErrNotLoggedIn
	}

	if err != nil {
		return Credentials{}, err
	}

	return parseCredentials(content, path)
}

func parseCredentials(content []byte, source string) (Credentials, error) {
	var creds Credentials

	if err := json.Unmarshal(content, &creds); err != nil {
		return Credentials{}, fmt.Errorf(
			"invalid credentials in %s: %w",
			source,
			err,
		)
	}

	return creds, nil
}

func SaveStoredCredentials(creds Credentials) (string, error) {
	secret, err := json.Marshal(creds)
	if err != nil {
		return "", err
	}

	path, err := credentialsFile()
	if err != nil {
		return "", err
	}

	if err := keyring.Set(
		keyringService,
		keyringUser,
		string(secret),
	); err == nil {
		_ = os.Remove(path)

		return "the system keyring", nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", err
	}

	if err := os.WriteFile(path, secret, 0o600); err != nil {
		return "", err
	}

	return path, nil
}

func DeleteStoredCredentials() error {
	keyringErr := keyring.Delete(keyringService, keyringUser)

	path, err := credentialsFile()
	if err != nil {
		return err
	}

	fileErr := os.Remove(path)

	if keyringErr != nil && errors.Is(fileErr, fs.ErrNotExist) {
		return ErrNotLoggedIn
	}

	if fileErr != nil && !errors.Is(fileErr, fs.ErrNotExist) {
		return fileErr
	}

	return nil
}

func NormalizeServer(server string) string {
	server = strings.TrimRight(strings.TrimSpace(server), "/")
	if !strings.Contains(server, "://") {
		server = "https://" + server
	}

	return server
}
