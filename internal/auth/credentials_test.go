package auth

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/zalando/go-keyring"
)

var testCreds = Credentials{
	Server:      "https://cloud.example.com",
	LoginName:   "alice",
	AppPassword: "secret",
	UserID:      "alice",
}

func setupConfigDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, ".config"))
	t.Setenv("AppData", filepath.Join(dir, "AppData"))

	path, err := credentialsFile()
	if err != nil {
		t.Fatal(err)
	}

	return path
}

func TestStoredCredentialsInKeyring(t *testing.T) {
	path := setupConfigDir(t)

	keyring.MockInit()

	location, err := SaveStoredCredentials(testCreds)
	if err != nil {
		t.Fatal(err)
	}

	if location != "the system keyring" {
		t.Fatalf("stored in %q, want the system keyring", location)
	}

	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("credentials file should not exist, got %v", err)
	}

	got, err := LoadStoredCredentials()
	if err != nil || got != testCreds {
		t.Fatalf("got %+v, %v", got, err)
	}

	if err := DeleteStoredCredentials(); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadStoredCredentials(); !errors.Is(err, ErrNotLoggedIn) {
		t.Fatalf("got %v, want ErrNotLoggedIn", err)
	}
}

func TestStoredCredentialsFallbackToFile(t *testing.T) {
	path := setupConfigDir(t)

	keyring.MockInitWithError(errors.New("no keyring available"))

	location, err := SaveStoredCredentials(testCreds)
	if err != nil {
		t.Fatal(err)
	}

	if location != path {
		t.Fatalf("stored in %q, want %q", location, path)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("file permissions are %o, want 600", perm)
	}

	got, err := LoadStoredCredentials()
	if err != nil || got != testCreds {
		t.Fatalf("got %+v, %v", got, err)
	}

	if err := DeleteStoredCredentials(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("credentials file should be removed, got %v", err)
	}

	if err := DeleteStoredCredentials(); !errors.Is(err, ErrNotLoggedIn) {
		t.Fatalf("got %v, want ErrNotLoggedIn", err)
	}
}
