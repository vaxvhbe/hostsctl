package cli

import (
	"github.com/spf13/cobra"
	"github.com/vaxvhbe/hostsctl/internal/profiles"
)

// buildCompletionCommand creates the completion command for generating shell completions.
func (c *CLI) buildCompletionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `Generate the autocompletion script for hostsctl for the specified shell.
See each sub-command's help for details on how to use the generated script.`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				return cmd.Root().GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				return cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
			}
			return nil
		},
	}

	cmd.AddCommand(c.buildBashCompletionCommand())
	cmd.AddCommand(c.buildZshCompletionCommand())
	cmd.AddCommand(c.buildFishCompletionCommand())
	cmd.AddCommand(c.buildPowershellCompletionCommand())

	return cmd
}

func (c *CLI) buildBashCompletionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "bash",
		Short: "Generate the autocompletion script for bash",
		Long: `Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(hostsctl completion bash)

To load completions for every new session, execute once:

Linux:

	hostsctl completion bash > /etc/bash_completion.d/hostsctl

macOS:

	hostsctl completion bash > /usr/local/etc/bash_completion.d/hostsctl

You will need to start a new shell for this setup to take effect.`,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenBashCompletion(cmd.OutOrStdout())
		},
	}
}

func (c *CLI) buildZshCompletionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "zsh",
		Short: "Generate the autocompletion script for zsh",
		Long: `Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it. You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(hostsctl completion zsh)

To load completions for every new session, execute once:

Linux:

	hostsctl completion zsh > "${fpath[1]}/_hostsctl"

macOS:

	hostsctl completion zsh > /usr/local/share/zsh/site-functions/_hostsctl

You will need to start a new shell for this setup to take effect.`,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenZshCompletion(cmd.OutOrStdout())
		},
	}
}

func (c *CLI) buildFishCompletionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "fish",
		Short: "Generate the autocompletion script for fish",
		Long: `Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	hostsctl completion fish | source

To load completions for every new session, execute once:

	hostsctl completion fish > ~/.config/fish/completions/hostsctl.fish

You will need to start a new shell for this setup to take effect.`,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
		},
	}
}

func (c *CLI) buildPowershellCompletionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "powershell",
		Short: "Generate the autocompletion script for powershell",
		Long: `Generate the autocompletion script for powershell.

To load completions in your current shell session:

	hostsctl completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.`,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
		},
	}
}

// setupCompletions sets up custom completions for various commands.
func (c *CLI) setupCompletions(rootCmd *cobra.Command) {
	// Setup profile name completion
	profileCompletion := func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		manager, err := profiles.NewManager()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		profileList, err := manager.ListProfiles()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		var names []string
		for _, profile := range profileList {
			names = append(names, profile.Name)
		}

		return names, cobra.ShellCompDirectiveNoFileComp
	}

	// Register profile name completions for relevant commands
	if profileCmd := findCommand(rootCmd, "profile"); profileCmd != nil {
		if applyCmd := findCommand(profileCmd, "apply"); applyCmd != nil {
			applyCmd.ValidArgsFunction = profileCompletion
		}
		if deleteCmd := findCommand(profileCmd, "delete"); deleteCmd != nil {
			deleteCmd.ValidArgsFunction = profileCompletion
		}
		if showCmd := findCommand(profileCmd, "show"); showCmd != nil {
			showCmd.ValidArgsFunction = profileCompletion
		}
		if diffCmd := findCommand(profileCmd, "diff"); diffCmd != nil {
			diffCmd.ValidArgsFunction = profileCompletion
		}
		if exportCmd := findCommand(profileCmd, "export"); exportCmd != nil {
			exportCmd.ValidArgsFunction = profileCompletion
		}
	}

	// Setup format completion for import/export commands
	formatCompletion := func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "yaml"}, cobra.ShellCompDirectiveNoFileComp
	}

	// Register format completions
	registerFlagCompletion(rootCmd, "import", "format", formatCompletion)
	registerFlagCompletion(rootCmd, "export", "format", formatCompletion)
	if profileCmd := findCommand(rootCmd, "profile"); profileCmd != nil {
		registerFlagCompletion(profileCmd, "import", "format", formatCompletion)
		registerFlagCompletion(profileCmd, "export", "format", formatCompletion)
	}

	// Setup status filter completion
	statusCompletion := func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"enabled", "disabled"}, cobra.ShellCompDirectiveNoFileComp
	}

	registerFlagCompletion(rootCmd, "list", "status-filter", statusCompletion)
}

// Helper function to find a command by name
func findCommand(parent *cobra.Command, name string) *cobra.Command {
	for _, cmd := range parent.Commands() {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}

// Helper function to register flag completion
func registerFlagCompletion(parent *cobra.Command, cmdName, flagName string, completion func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective)) {
	if cmd := findCommand(parent, cmdName); cmd != nil {
		_ = cmd.RegisterFlagCompletionFunc(flagName, completion)
	}
}
