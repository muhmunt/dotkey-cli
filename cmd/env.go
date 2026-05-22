package cmd

import (
	"dotkey-cli/internal/config"
	"dotkey-cli/internal/output"
	"fmt"

	"github.com/spf13/cobra"
)

func newEnvsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "envs",
		Short: "List environments in the active project",
		RunE: func(cmd *cobra.Command, args []string) error {
			checkAuth()
			projectID, projectName := requireProject()

			envs, err := client.ListEnvironments(projectID)
			if err != nil {
				return err
			}

			fmt.Printf("\n  Environments — %s (%d)\n\n", output.Bold(projectName), len(envs))

			rows := make([][]string, len(envs))
			for i, e := range rows {
				_ = e
				env := envs[i]
				active := ""
				if env.ID == cfg.CurrentEnvID {
					active = output.Dim("← active")
				}
				rows[i] = []string{env.Name, output.Dim(env.ID[:8] + "…"), active}
			}
			output.Table([]string{"name", "id", ""}, rows)
			fmt.Println()
			return nil
		},
	}
}

func newEnvCmd() *cobra.Command {
	envCmd := &cobra.Command{
		Use:   "env",
		Short: "Manage environments",
	}
	envCmd.AddCommand(
		envCreateCmd(),
		envUseCmd(),
		envDeleteCmd(),
	)
	return envCmd
}

func envCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkAuth()
			projectID, _ := requireProject()

			env, err := client.CreateEnvironment(projectID, args[0])
			if err != nil {
				return err
			}

			output.Success("Environment created: %s", output.Bold(env.Name))
			return nil
		},
	}
}

func envUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <name>",
		Short: "Set the active environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkAuth()
			projectID, _ := requireProject()

			envs, err := client.ListEnvironments(projectID)
			if err != nil {
				return err
			}
			for _, e := range envs {
				if e.Name == args[0] || e.ID == args[0] {
					cfg.CurrentEnvID = e.ID
					cfg.CurrentEnvName = e.Name
					if err := config.Save(cfg); err != nil {
						return err
					}
					output.Success("Now using: %s", output.Bold(e.Name))
					return nil
				}
			}
			return fmt.Errorf("environment %q not found.\nRun `dotkey envs` to see available environments", args[0])
		},
	}
}

func envDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete an environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkAuth()
			projectID, _ := requireProject()

			envID, err := resolveEnvByName(projectID, args[0])
			if err != nil {
				return err
			}

			if !output.Confirm(fmt.Sprintf("Delete environment %q and all its variables?", args[0])) {
				output.Info("Cancelled")
				return nil
			}

			if err := client.DeleteEnvironment(projectID, envID); err != nil {
				return err
			}

			// clear active env if it was the deleted one
			if cfg.CurrentEnvID == envID {
				cfg.CurrentEnvID = ""
				cfg.CurrentEnvName = ""
				config.Save(cfg) //nolint:errcheck
			}

			output.Success("Environment %q deleted", args[0])
			return nil
		},
	}
}
