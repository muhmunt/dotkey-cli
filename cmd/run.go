package cmd

import (
	"dotkey-cli/internal/dotenv"
	"dotkey-cli/internal/output"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

func newRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run [environment] -- <command> [args...]",
		Short: "Run a command with secrets injected as environment variables",
		Long: `Pulls variables for the active environment and runs the given command
with those variables injected into its environment. Secrets are never
written to disk.

Examples:
  dotkey run -- npm run dev
  dotkey run production -- node server.js
  dotkey run -p my-app -e staging -- pytest`,
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			checkAuth()
			projectID, projectName := requireProject()

			// Split args at -- into [env-name?] and [command args...]
			dashIdx := cmd.ArgsLenAtDash()
			var envArg string
			var cmdArgs []string

			switch {
			case dashIdx == -1:
				// No -- at all: everything is the command
				cmdArgs = args
			case dashIdx == 0:
				// dotkey run -- cmd args
				cmdArgs = args
			default:
				// dotkey run [env] -- cmd args
				before := args[:dashIdx]
				cmdArgs = args[dashIdx:]
				if len(before) > 0 {
					envArg = before[0]
				}
			}

			if len(cmdArgs) == 0 {
				return fmt.Errorf("no command specified\n\nUsage: dotkey run -- <command> [args...]")
			}

			// Resolve environment
			var envID, envName string
			if envArg != "" {
				var err error
				envID, err = resolveEnvByName(projectID, envArg)
				if err != nil {
					return err
				}
				envName = envArg
			} else {
				envID, envName = requireEnv(projectID)
			}

			printContext(projectName, envName)
			output.Info("Injecting %s secrets…", envName)

			content, err := client.Export(projectID, envID)
			if err != nil {
				return err
			}

			secretVars, err := dotenv.ParseString(content)
			if err != nil {
				return err
			}

			// Build env: inherit current process env, then overlay secrets
			// (secrets take precedence over existing env vars with the same name)
			env := os.Environ()
			for k, v := range secretVars {
				env = append(env, k+"="+v)
			}

			fmt.Printf("  %s  %d variables injected\n\n", output.Dim("✓"), len(secretVars))

			proc := exec.Command(cmdArgs[0], cmdArgs[1:]...)
			proc.Env = env
			proc.Stdin = os.Stdin
			proc.Stdout = os.Stdout
			proc.Stderr = os.Stderr

			// Forward SIGINT/SIGTERM to child so Ctrl-C works cleanly
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				if sig, ok := <-sigs; ok {
					if proc.Process != nil {
						proc.Process.Signal(sig) //nolint:errcheck
					}
				}
			}()

			if err := proc.Run(); err != nil {
				signal.Stop(sigs)
				close(sigs)
				var exitErr *exec.ExitError
				if errors.As(err, &exitErr) {
					os.Exit(exitErr.ExitCode())
				}
				return err
			}

			signal.Stop(sigs)
			close(sigs)
			return nil
		},
	}
}
