package cmd

import (
	"dotkey-cli/internal/api"
	"dotkey-cli/internal/config"
	"dotkey-cli/internal/output"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	cfg    *config.Config
	client *api.Client

	flagAPI   string
	flagWeb   string
	flagToken string
	flagQuiet bool

	flagProject string
	flagEnv     string
)

var rootCmd = &cobra.Command{
	Use:     "dotkey",
	Short:   "dotkey — manage environment variables from the terminal",
	Long:    `dotkey lets you pull, push, and manage environment variables\nwithout opening a browser or sharing .env files manually.`,
	Version: config.Version,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		output.Fatal("%s", err.Error())
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&flagAPI, "api", "", "override API base URL")
	rootCmd.PersistentFlags().StringVar(&flagWeb, "web", "", "override web dashboard URL")
	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "override auth token (useful in CI)")
	rootCmd.PersistentFlags().BoolVarP(&flagQuiet, "quiet", "q", false, "suppress output except errors")
	rootCmd.PersistentFlags().StringVarP(&flagProject, "project", "p", "", "override active project (name or ID)")
	rootCmd.PersistentFlags().StringVarP(&flagEnv, "env", "e", "", "override active environment (name or ID)")

	rootCmd.AddCommand(
		newAuthCmd(),
		newLogoutCmd(),
		newWhoamiCmd(),
		newProjectsCmd(),
		newProjectCmd(),
		newEnvsCmd(),
		newEnvCmd(),
		newPullCmd(),
		newPushCmd(),
		newAddCmd(),
		newRemoveCmd(),
		newGetCmd(),
		newListCmd(),
		newSyncCmd(),
		newDiffCmd(),
		newHistoryCmd(),
		newRollbackCmd(),
	)
}

func initConfig() {
	cfg = config.Load()

	if flagAPI != "" {
		cfg.APIURL = flagAPI
	}
	if flagWeb != "" {
		cfg.WebURL = flagWeb
	}
	if flagToken != "" {
		cfg.Token = flagToken
	}
	if flagQuiet {
		output.Quiet = true
	}

	client = api.NewClient(cfg.APIURL, cfg.Token)
}

// ── Context helpers ──────────────────────────────────────────────────────────

func requireProject() (id, name string) {
	if flagProject != "" {
		projects, err := client.ListProjects()
		if err != nil {
			output.Fatal("%s", err)
		}
		for _, p := range projects {
			if p.Name == flagProject || p.ID == flagProject {
				return p.ID, p.Name
			}
		}
		output.Fatal("project %q not found.\nRun `dotkey projects` to see available projects", flagProject)
	}

	if lf := config.LoadLocalFile(); lf != nil && lf.Project != "" {
		projects, err := client.ListProjects()
		if err != nil {
			output.Fatal("%s", err)
		}
		for _, p := range projects {
			if p.Name == lf.Project {
				return p.ID, p.Name
			}
		}
	}

	if cfg.CurrentProjectID == "" {
		output.Fatal("no project selected.\nRun `dotkey project use <name>` to select a project")
	}
	return cfg.CurrentProjectID, cfg.CurrentProjectName
}

func requireEnv(projectID string) (id, name string) {
	if flagEnv != "" {
		envs, err := client.ListEnvironments(projectID)
		if err != nil {
			output.Fatal("%s", err)
		}
		for _, e := range envs {
			if e.Name == flagEnv || e.ID == flagEnv {
				return e.ID, e.Name
			}
		}
		output.Fatal("environment %q not found.\nRun `dotkey envs` to see available environments", flagEnv)
	}

	if lf := config.LoadLocalFile(); lf != nil && lf.Environment != "" {
		envs, err := client.ListEnvironments(projectID)
		if err != nil {
			output.Fatal("%s", err)
		}
		for _, e := range envs {
			if e.Name == lf.Environment {
				return e.ID, e.Name
			}
		}
	}

	if cfg.CurrentEnvID == "" {
		output.Fatal("no environment selected.\nRun `dotkey env use <name>` to select an environment")
	}
	return cfg.CurrentEnvID, cfg.CurrentEnvName
}

func resolveEnvByName(projectID, nameOrID string) (string, error) {
	envs, err := client.ListEnvironments(projectID)
	if err != nil {
		return "", err
	}
	for _, e := range envs {
		if e.Name == nameOrID || e.ID == nameOrID {
			return e.ID, nil
		}
	}
	return "", fmt.Errorf("environment %q not found", nameOrID)
}

func checkAuth() {
	if cfg.Token == "" {
		output.Fatal("not logged in. Run `dotkey login` first")
	}
}

func printContext(projectName, envName string) {
	fmt.Printf("\n  %s  %s\n  %s  %s\n\n",
		output.Dim("Project"), output.Bold(projectName),
		output.Dim("Env    "), output.Bold(envName),
	)
}
