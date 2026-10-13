package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/plivox/nextcloud-cli/internal/auth"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login <server>",
	Short: "Log in through the browser and store an app password (system keyring, or a file in the user config dir)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		server := auth.NormalizeServer(args[0])

		flow, err := auth.StartLoginFlow(server)
		if err != nil {
			return err
		}

		fmt.Fprintf(
			os.Stderr,
			"Open this URL in your browser to log in:\n\n  %s\n\n",
			flow.LoginURL,
		)
		_ = openBrowser(flow.LoginURL)

		fmt.Fprintln(os.Stderr, "Waiting for authorization...")

		creds, err := flow.Wait(20 * time.Minute)
		if err != nil {
			return err
		}

		userID, err := creds.Client().CurrentUserID()
		if err != nil {
			return err
		}

		creds.UserID = userID

		location, err := auth.SaveStoredCredentials(creds)
		if err != nil {
			return fmt.Errorf("saving credentials: %w", err)
		}

		fmt.Fprintf(os.Stdout, "Logged in to %s as %s\n", creds.Server, userID)
		fmt.Fprintf(os.Stderr, "Credentials stored in %s\n", location)

		if auth.CredentialsFromEnv() {
			fmt.Fprintln(
				os.Stderr,
				"Warning: NEXTCLOUD_CLI_* variables are set (environment or .env) and take precedence over the stored credentials.",
			)
		}

		return nil
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Revoke the stored app password and remove it from this machine",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		creds, err := auth.LoadStoredCredentials()
		if errors.Is(err, auth.ErrNotLoggedIn) {
			fmt.Fprintln(os.Stdout, "Not logged in.")
			return nil
		}

		if err != nil {
			return err
		}

		if err := creds.Client().RevokeAppPassword(); err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Warning: %v (removing it locally anyway)\n",
				err,
			)
		}

		if err := auth.DeleteStoredCredentials(); err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "Logged out from %s\n", creds.Server)

		return nil
	},
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "windows":
		return exec.Command(
			"rundll32",
			"url.dll,FileProtocolHandler",
			url,
		).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
}
