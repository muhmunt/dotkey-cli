package cmd

import (
	"dotkey-cli/internal/api"
	"dotkey-cli/internal/config"
	"dotkey-cli/internal/output"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

func newAuthCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate via browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			dc, err := client.DeviceCode()
			if err != nil {
				return fmt.Errorf("failed to start login: %w", err)
			}

			activateURL := cfg.WebURL + "/activate?code=" + dc.UserCode

			fmt.Println()
			output.Info("Opening browser for authentication...")
			fmt.Printf("\n  If the browser did not open, visit:\n  %s\n", output.Bold(activateURL))
			fmt.Printf("\n  Enter this code: %s\n\n", output.Bold(dc.UserCode))

			openBrowser(activateURL)
			output.Info("Waiting for approval...")

			token, err := pollForToken(dc.DeviceCode, dc.ExpiresIn)
			if err != nil {
				return err
			}

			authedClient := api.NewClient(cfg.APIURL, token)
			user, err := authedClient.Me()
			if err != nil {
				return fmt.Errorf("authenticated but failed to fetch user info: %w", err)
			}

			cfg.Token = token
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save credentials: %w", err)
			}

			fmt.Println()
			output.Success("Authenticated as %s (%s)", output.Bold(user.Name), output.Dim(user.Email))
			fmt.Println()
			return nil
		},
	}
}

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove local credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.Token = ""
			cfg.CurrentProjectID = ""
			cfg.CurrentProjectName = ""
			cfg.CurrentEnvID = ""
			cfg.CurrentEnvName = ""
			if err := config.Save(cfg); err != nil {
				return err
			}
			output.Success("Logged out")
			return nil
		},
	}
}

func newWhoamiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show current logged-in user",
		RunE: func(cmd *cobra.Command, args []string) error {
			checkAuth()
			user, err := client.Me()
			if err != nil {
				return err
			}
			fmt.Println()
			fmt.Printf("  %s  %s\n", output.Dim("Name "), output.Bold(user.Name))
			fmt.Printf("  %s  %s\n", output.Dim("Email"), user.Email)
			fmt.Printf("  %s  %s\n", output.Dim("ID   "), output.Dim(user.ID))
			fmt.Println()
			return nil
		},
	}
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func pollForToken(deviceCode string, expiresIn int) (string, error) {
	deadline := time.Now().Add(time.Duration(expiresIn) * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(3 * time.Second)
		result, err := client.DevicePoll(deviceCode)
		if err != nil {
			continue
		}
		if result.Status == "pending" {
			continue
		}
		if result.Token != "" {
			return result.Token, nil
		}
	}
	return "", fmt.Errorf("login timed out. Run `dotkey login` to try again")
}

func openBrowser(url string) {
	var c *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		c = exec.Command("open", url)
	case "linux":
		c = exec.Command("xdg-open", url)
	case "windows":
		c = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}
	c.Start() //nolint:errcheck
}
