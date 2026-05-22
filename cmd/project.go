package cmd

import (
	"dotkey-cli/internal/config"
	"dotkey-cli/internal/output"
	"fmt"

	"github.com/spf13/cobra"
)

func newProjectsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "projects",
		Short: "List all your projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			checkAuth()
			projects, err := client.ListProjects()
			if err != nil {
				return err
			}

			fmt.Printf("\n  Projects (%d)\n\n", len(projects))

			rows := make([][]string, len(projects))
			for i, p := range projects {
				active := ""
				if p.ID == cfg.CurrentProjectID {
					active = output.Dim("← active")
				}
				rows[i] = []string{p.Name, output.Dim(p.ID[:8] + "…"), active}
			}
			output.Table([]string{"name", "id", ""}, rows)
			fmt.Println()
			return nil
		},
	}
}

func newProjectCmd() *cobra.Command {
	projectCmd := &cobra.Command{
		Use:   "project",
		Short: "Manage projects",
	}

	projectCmd.AddCommand(
		projectCreateCmd(),
		projectUseCmd(),
		projectInfoCmd(),
	)
	return projectCmd
}

func projectCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create a new project",
		RunE: func(cmd *cobra.Command, args []string) error {
			checkAuth()

			name := output.Prompt("Project name")
			if name == "" {
				return fmt.Errorf("project name cannot be empty")
			}
			desc := output.Prompt("Description (optional)")

			p, err := client.CreateProject(name, desc)
			if err != nil {
				return err
			}

			// auto-select the newly created project
			cfg.CurrentProjectID = p.ID
			cfg.CurrentProjectName = p.Name
			cfg.CurrentEnvID = ""
			cfg.CurrentEnvName = ""
			config.Save(cfg) //nolint:errcheck

			fmt.Println()
			output.Success("Project created: %s", output.Bold(p.Name))
			output.Success("Now using: %s", output.Bold(p.Name))
			fmt.Println()
			return nil
		},
	}
}

func projectUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <name>",
		Short: "Set the active project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			checkAuth()
			projects, err := client.ListProjects()
			if err != nil {
				return err
			}
			for _, p := range projects {
				if p.Name == args[0] || p.ID == args[0] {
					cfg.CurrentProjectID = p.ID
					cfg.CurrentProjectName = p.Name
					cfg.CurrentEnvID = ""
					cfg.CurrentEnvName = ""
					if err := config.Save(cfg); err != nil {
						return err
					}
					output.Success("Now using: %s", output.Bold(p.Name))
					return nil
				}
			}
			return fmt.Errorf("project %q not found.\nRun `dotkey projects` to see available projects", args[0])
		},
	}
}

func projectInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show active project details",
		RunE: func(cmd *cobra.Command, args []string) error {
			checkAuth()
			projectID, projectName := requireProject()

			members, err := client.ListMembers(projectID)
			if err != nil {
				return err
			}

			envs, err := client.ListEnvironments(projectID)
			if err != nil {
				return err
			}

			fmt.Println()
			fmt.Printf("  %s  %s\n", output.Dim("Project "), output.Bold(projectName))
			fmt.Printf("  %s  %s\n\n", output.Dim("ID      "), output.Dim(projectID))

			fmt.Printf("  %s\n\n", output.Bold("Environments"))
			for _, e := range envs {
				fmt.Printf("    %s\n", e.Name)
			}

			fmt.Printf("\n  %s\n\n", output.Bold("Members"))
			for _, m := range members {
				name := output.Dim("(unknown)")
				if m.User != nil {
					name = m.User.Name
				}
				fmt.Printf("    %-20s  %s\n", name, output.Dim(m.Role))
			}
			fmt.Println()
			return nil
		},
	}
}
