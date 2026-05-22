package cmd

import (
	"dotkey-cli/internal/output"
	"fmt"

	"github.com/spf13/cobra"
)

func newDiffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diff <env-a> <env-b>",
		Short: "Compare two environments",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkAuth()
			projectID, projectName := requireProject()

			envAID, err := resolveEnvByName(projectID, args[0])
			if err != nil {
				return err
			}
			envBID, err := resolveEnvByName(projectID, args[1])
			if err != nil {
				return err
			}

			entries, err := client.Diff(projectID, envAID, envBID)
			if err != nil {
				return err
			}

			fmt.Printf("\n  Diff — %s / %s → %s\n\n",
				output.Bold(projectName),
				output.Bold(args[0]),
				output.Bold(args[1]),
			)

			rows := make([][]string, len(entries))
			for i, e := range entries {
				rows[i] = []string{e.Key, e.Status}
			}
			output.DiffTable([]string{"key", "status"}, rows, 1)

			// summary
			var changed, missingB, missingA, same int
			for _, e := range entries {
				switch e.Status {
				case "changed":
					changed++
				case "missing_in_b":
					missingB++
				case "missing_in_a":
					missingA++
				case "same":
					same++
				}
			}

			fmt.Printf("\n  %s  %s  %s  %s\n\n",
				output.Dim(fmt.Sprintf("%d changed", changed)),
				output.Dim(fmt.Sprintf("%d missing in %s", missingB, args[1])),
				output.Dim(fmt.Sprintf("%d missing in %s", missingA, args[0])),
				output.Dim(fmt.Sprintf("%d same", same)),
			)
			return nil
		},
	}
}
