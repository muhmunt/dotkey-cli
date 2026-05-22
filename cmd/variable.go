package cmd

import (
	"dotkey-cli/internal/dotenv"
	"dotkey-cli/internal/output"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newPullCmd() *cobra.Command {
	var outFile string
	cmd := &cobra.Command{
		Use:   "pull [environment]",
		Short: "Pull env vars from server and write .env file",
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

			printContext(projectName, envName)

			content, err := client.Export(projectID, envID)
			if err != nil {
				return err
			}

			vars, err := dotenv.ParseString(content)
			if err != nil {
				return err
			}

			dest := outFile
			if dest == "" {
				dest = ".env"
			}

			if err := dotenv.Write(dest, dotenv.Serialize(vars)); err != nil {
				return fmt.Errorf("cannot write %s: %w", dest, err)
			}

			output.Info("Pulling %d variables...", len(vars))
			for k := range vars {
				fmt.Printf("  %s  %s\n", output.Dim("✓"), k)
			}
			fmt.Println()
			output.Success("Written to %s", dest)
			fmt.Println()
			return nil
		},
	}
	cmd.Flags().StringVarP(&outFile, "output", "o", "", "output file (default: .env)")
	return cmd
}

func newPushCmd() *cobra.Command {
	var inFile string
	cmd := &cobra.Command{
		Use:   "push [environment]",
		Short: "Push local .env file to server",
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

			src := inFile
			if src == "" {
				src = ".env"
			}

			vars, err := dotenv.Parse(src)
			if err != nil {
				return fmt.Errorf("cannot read %s: %w", src, err)
			}

			printContext(projectName, envName)
			output.Info("Pushing %d variables...", len(vars))

			result, err := client.Import(projectID, envID, dotenv.ToLines(vars))
			if err != nil {
				return err
			}

			fmt.Println()
			output.Success("%d variables synced to %s", result.Synced, envName)
			fmt.Println()
			return nil
		},
	}
	cmd.Flags().StringVarP(&inFile, "file", "f", "", "input file (default: .env)")
	return cmd
}

func newAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add KEY=VALUE",
		Short: "Add or update a single variable",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkAuth()

			parts := strings.SplitN(args[0], "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("expected KEY=VALUE format, got %q", args[0])
			}
			key, value := strings.TrimSpace(parts[0]), parts[1]

			projectID, _ := requireProject()
			envID, envName := requireEnv(projectID)

			// try import (upsert) for simplicity
			result, err := client.Import(projectID, envID, key+"="+value)
			if err != nil {
				return err
			}

			if result.Synced > 0 {
				output.Success("%s set in %s", output.Bold(key), envName)
			} else {
				output.Warn("%s was not changed (value may be identical)", key)
			}
			return nil
		},
	}
}

func newRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove KEY",
		Short: "Delete a variable",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkAuth()
			key := args[0]
			projectID, _ := requireProject()
			envID, envName := requireEnv(projectID)

			vars, err := client.ListVariables(projectID, envID)
			if err != nil {
				return err
			}

			var varID string
			for _, v := range vars {
				if v.Key == key {
					varID = v.ID
					break
				}
			}
			if varID == "" {
				return fmt.Errorf("variable %q not found in %s", key, envName)
			}

			if err := client.DeleteVariable(projectID, envID, varID); err != nil {
				return err
			}

			output.Success("%s removed from %s", output.Bold(key), envName)
			return nil
		},
	}
}

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get KEY",
		Short: "Show the decrypted value of a variable",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkAuth()
			key := args[0]
			projectID, _ := requireProject()
			envID, _ := requireEnv(projectID)

			content, err := client.Export(projectID, envID)
			if err != nil {
				return err
			}

			vars, err := dotenv.ParseString(content)
			if err != nil {
				return err
			}

			val, ok := vars[key]
			if !ok {
				return fmt.Errorf("variable %q not found", key)
			}

			fmt.Printf("%s=%s\n", key, val)
			return nil
		},
	}
}

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list [environment]",
		Short: "List variable keys in the active environment",
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

			vars, err := client.ListVariables(projectID, envID)
			if err != nil {
				return err
			}

			fmt.Printf("\n  Variables — %s / %s (%d)\n\n", output.Bold(projectName), output.Bold(envName), len(vars))

			rows := make([][]string, len(vars))
			for i, v := range vars {
				updated := v.UpdatedAt
				if len(updated) > 10 {
					updated = updated[:10]
				}
				rows[i] = []string{v.Key, output.Dim(updated)}
			}
			output.Table([]string{"key", "updated"}, rows)
			fmt.Println()
			return nil
		},
	}
}

func newSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Two-way sync: push local .env then pull from server",
		RunE: func(cmd *cobra.Command, args []string) error {
			checkAuth()
			projectID, projectName := requireProject()
			envID, envName := requireEnv(projectID)

			printContext(projectName, envName)

			// step 1: push local .env
			vars, err := dotenv.Parse(".env")
			if err != nil {
				output.Warn("No local .env found, skipping push")
			} else {
				output.Info("Pushing %d local variables...", len(vars))
				if _, err := client.Import(projectID, envID, dotenv.ToLines(vars)); err != nil {
					return fmt.Errorf("push failed: %w", err)
				}
				output.Success("Local vars pushed")
			}

			// step 2: pull from server
			output.Info("Pulling from server...")
			content, err := client.Export(projectID, envID)
			if err != nil {
				return fmt.Errorf("pull failed: %w", err)
			}

			serverVars, _ := dotenv.ParseString(content)
			if err := dotenv.Write(".env", dotenv.Serialize(serverVars)); err != nil {
				return err
			}

			output.Success("%d variables written to .env", len(serverVars))
			fmt.Println()
			return nil
		},
	}
}
