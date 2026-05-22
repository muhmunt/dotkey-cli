package cmd

import (
	"dotkey-cli/internal/output"
	"fmt"

	"github.com/spf13/cobra"
)

func newHistoryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "history [environment]",
		Short: "Show variable change history",
		RunE: func(cmd *cobra.Command, args []string) error {
			checkAuth()
			projectID, projectName := requireProject()

			var envID, envName string
			if len(args) > 0 {
				var err error
				envID, err = resolveEnvByName(projectID, args[0])
				if err != nil {
					return err
				}
				envName = args[0]
			} else {
				envID, envName = requireEnv(projectID)
			}

			versions, err := client.History(projectID, envID)
			if err != nil {
				return err
			}

			fmt.Printf("\n  History — %s / %s\n\n", output.Bold(projectName), output.Bold(envName))

			rows := make([][]string, len(versions))
			for i, v := range versions {
				actor := output.Dim("unknown")
				if v.Actor != nil {
					actor = v.Actor.Name
				}
				when := v.CreatedAt
				if len(when) > 10 {
					when = when[:10]
				}
				shortID := v.ID
				if len(shortID) > 8 {
					shortID = shortID[:8] + "…"
				}
				rows[i] = []string{shortID, v.Key, v.Action, actor, output.Dim(when)}
			}
			output.Table([]string{"version", "key", "action", "by", "when"}, rows)
			fmt.Println()
			return nil
		},
	}
}

func newRollbackCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rollback <version-id>",
		Short: "Restore a variable to a previous value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkAuth()
			projectID, _ := requireProject()
			envID, envName := requireEnv(projectID)

			versionID := args[0]

			if !output.Confirm(fmt.Sprintf("Roll back version %q in %s?", versionID[:min(8, len(versionID))], envName)) {
				output.Info("Cancelled")
				return nil
			}

			if err := client.Rollback(projectID, envID, versionID); err != nil {
				return err
			}

			output.Success("Rolled back to version %s", output.Dim(versionID[:min(8, len(versionID))]))
			return nil
		},
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
