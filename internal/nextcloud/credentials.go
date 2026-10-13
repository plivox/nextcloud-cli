package nextcloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

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

	return LoadKeyringCredentials()
}

func CredentialsFromEnv() bool {
	return os.Getenv("NEXTCLOUD_CLI_BASE_URL") != "" ||
		os.Getenv("NEXTCLOUD_CLI_USERNAME") != "" ||
		os.Getenv("NEXTCLOUD_CLI_PASSWORD") != ""
}

func LoadKeyringCredentials() (Credentials, error) {
	secret, err := keyring.Get(keyringService, keyringUser)
	if errors.Is(err, keyring.ErrNotFound) {
		return Credentials{}, ErrNotLoggedIn
	}

	if err != nil {
		return Credentials{}, fmt.Errorf("reading the keyring: %w", err)
	}

	var creds Credentials

	if err := json.Unmarshal([]byte(secret), &creds); err != nil {
		return Credentials{}, fmt.Errorf(
			"invalid credentials in the keyring: %w",
			err,
		)
	}

	return creds, nil
}

func SaveKeyringCredentials(creds Credentials) error {
	secret, err := json.Marshal(creds)
	if err != nil {
		return err
	}

	return keyring.Set(keyringService, keyringUser, string(secret))
}

func DeleteKeyringCredentials() error {
	err := keyring.Delete(keyringService, keyringUser)
	if errors.Is(err, keyring.ErrNotFound) {
		return ErrNotLoggedIn
	}

	return err
}

func NormalizeServer(server string) string {
	server = strings.TrimRight(strings.TrimSpace(server), "/")
	if !strings.Contains(server, "://") {
		server = "https://" + server
	}

	return server
}
