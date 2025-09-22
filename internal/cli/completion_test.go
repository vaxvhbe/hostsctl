package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestCLI_buildCompletionCommand(t *testing.T) {
	cli := NewCLI()
	cmd := cli.buildCompletionCommand()

	if cmd == nil {
		t.Error("buildCompletionCommand() returned nil")
		return
	}

	if cmd.Use != "completion [bash|zsh|fish|powershell]" {
		t.Errorf("buildCompletionCommand() Use = %v, want %v", cmd.Use, "completion [bash|zsh|fish|powershell]")
	}

	// Test valid arguments
	expectedArgs := []string{"bash", "zsh", "fish", "powershell"}
	if len(cmd.ValidArgs) != len(expectedArgs) {
		t.Errorf("buildCompletionCommand() ValidArgs length = %d, want %d", len(cmd.ValidArgs), len(expectedArgs))
	}

	for i, arg := range expectedArgs {
		if i >= len(cmd.ValidArgs) || cmd.ValidArgs[i] != arg {
			t.Errorf("buildCompletionCommand() ValidArgs[%d] = %v, want %v", i, cmd.ValidArgs[i], arg)
		}
	}

	// Test subcommands
	subcommands := make(map[string]bool)
	for _, subcmd := range cmd.Commands() {
		subcommands[subcmd.Use] = true
	}

	expectedSubcommands := []string{"bash", "zsh", "fish", "powershell"}
	for _, expected := range expectedSubcommands {
		if !subcommands[expected] {
			t.Errorf("buildCompletionCommand() missing subcommand: %s", expected)
		}
	}
}

func TestCLI_buildBashCompletionCommand(t *testing.T) {
	cli := NewCLI()
	cmd := cli.buildBashCompletionCommand()

	if cmd == nil {
		t.Error("buildBashCompletionCommand() returned nil")
		return
	}

	if cmd.Use != "bash" {
		t.Errorf("buildBashCompletionCommand() Use = %v, want %v", cmd.Use, "bash")
	}

	if cmd.Short == "" {
		t.Error("buildBashCompletionCommand() should have a short description")
	}

	if cmd.Long == "" {
		t.Error("buildBashCompletionCommand() should have a long description")
	}

	if !cmd.DisableFlagsInUseLine {
		t.Error("buildBashCompletionCommand() should disable flags in usage line")
	}
}

func TestCLI_buildZshCompletionCommand(t *testing.T) {
	cli := NewCLI()
	cmd := cli.buildZshCompletionCommand()

	if cmd == nil {
		t.Error("buildZshCompletionCommand() returned nil")
		return
	}

	if cmd.Use != "zsh" {
		t.Errorf("buildZshCompletionCommand() Use = %v, want %v", cmd.Use, "zsh")
	}

	if cmd.Short == "" {
		t.Error("buildZshCompletionCommand() should have a short description")
	}

	if cmd.Long == "" {
		t.Error("buildZshCompletionCommand() should have a long description")
	}

	if !cmd.DisableFlagsInUseLine {
		t.Error("buildZshCompletionCommand() should disable flags in usage line")
	}
}

func TestCLI_buildFishCompletionCommand(t *testing.T) {
	cli := NewCLI()
	cmd := cli.buildFishCompletionCommand()

	if cmd == nil {
		t.Error("buildFishCompletionCommand() returned nil")
		return
	}

	if cmd.Use != "fish" {
		t.Errorf("buildFishCompletionCommand() Use = %v, want %v", cmd.Use, "fish")
	}

	if cmd.Short == "" {
		t.Error("buildFishCompletionCommand() should have a short description")
	}

	if cmd.Long == "" {
		t.Error("buildFishCompletionCommand() should have a long description")
	}

	if !cmd.DisableFlagsInUseLine {
		t.Error("buildFishCompletionCommand() should disable flags in usage line")
	}
}

func TestCLI_buildPowershellCompletionCommand(t *testing.T) {
	cli := NewCLI()
	cmd := cli.buildPowershellCompletionCommand()

	if cmd == nil {
		t.Error("buildPowershellCompletionCommand() returned nil")
		return
	}

	if cmd.Use != "powershell" {
		t.Errorf("buildPowershellCompletionCommand() Use = %v, want %v", cmd.Use, "powershell")
	}

	if cmd.Short == "" {
		t.Error("buildPowershellCompletionCommand() should have a short description")
	}

	if cmd.Long == "" {
		t.Error("buildPowershellCompletionCommand() should have a long description")
	}

	if !cmd.DisableFlagsInUseLine {
		t.Error("buildPowershellCompletionCommand() should disable flags in usage line")
	}
}

func TestFindCommand(t *testing.T) {
	// Create a test command structure
	parent := &cobra.Command{
		Use: "parent",
	}

	child1 := &cobra.Command{
		Use: "child1",
	}

	child2 := &cobra.Command{
		Use: "child2",
	}

	parent.AddCommand(child1)
	parent.AddCommand(child2)

	// Test finding existing commands
	found := findCommand(parent, "child1")
	if found == nil {
		t.Error("findCommand should find existing command")
	} else if found.Use != "child1" {
		t.Errorf("findCommand returned wrong command: %s", found.Use)
	}

	found = findCommand(parent, "child2")
	if found == nil {
		t.Error("findCommand should find existing command")
	} else if found.Use != "child2" {
		t.Errorf("findCommand returned wrong command: %s", found.Use)
	}

	// Test finding non-existent command
	found = findCommand(parent, "nonexistent")
	if found != nil {
		t.Error("findCommand should return nil for non-existent command")
	}
}

func TestRegisterFlagCompletion(t *testing.T) {
	// Create a test command structure
	parent := &cobra.Command{
		Use: "parent",
	}

	child := &cobra.Command{
		Use: "child",
	}
	child.Flags().String("test-flag", "", "test flag")

	parent.AddCommand(child)

	// Test completion function
	completionFunc := func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"option1", "option2"}, cobra.ShellCompDirectiveNoFileComp
	}

	// This should not panic
	registerFlagCompletion(parent, "child", "test-flag", completionFunc)

	// Test with non-existent command (should not panic)
	registerFlagCompletion(parent, "nonexistent", "test-flag", completionFunc)

	// Test with non-existent flag (should not panic)
	registerFlagCompletion(parent, "child", "nonexistent-flag", completionFunc)
}

func TestCLI_setupCompletions(t *testing.T) {
	cli := NewCLI()
	rootCmd := cli.buildRootCommand()

	// This should not panic
	cli.setupCompletions(rootCmd)

	// Verify that completion setup doesn't break the command structure
	if rootCmd == nil {
		t.Error("setupCompletions should not break root command")
	}

	// Test that commands still exist after setup
	profileCmd := findCommand(rootCmd, "profile")
	if profileCmd == nil {
		t.Error("setupCompletions should not remove profile command")
	}

	listCmd := findCommand(rootCmd, "list")
	if listCmd == nil {
		t.Error("setupCompletions should not remove list command")
	}
}

func TestCompletionFunctions(t *testing.T) {
	// Test that completion functions have the correct signature
	formatCompletion := func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "yaml"}, cobra.ShellCompDirectiveNoFileComp
	}

	statusCompletion := func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"enabled", "disabled"}, cobra.ShellCompDirectiveNoFileComp
	}

	// Test format completion
	options, directive := formatCompletion(nil, nil, "")
	if len(options) != 2 {
		t.Errorf("formatCompletion should return 2 options, got %d", len(options))
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Error("formatCompletion should return NoFileComp directive")
	}

	// Test status completion
	options, directive = statusCompletion(nil, nil, "")
	if len(options) != 2 {
		t.Errorf("statusCompletion should return 2 options, got %d", len(options))
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Error("statusCompletion should return NoFileComp directive")
	}
}

func TestCLI_completionIntegration(t *testing.T) {
	cli := NewCLI()

	// Build the complete command tree
	rootCmd := cli.buildRootCommand()

	// Setup completions
	cli.setupCompletions(rootCmd)

	// Test that the command tree is still valid
	if rootCmd == nil {
		t.Error("Root command should not be nil after completion setup")
	}

	// Test that completion command exists
	completionCmd := findCommand(rootCmd, "completion")
	if completionCmd == nil {
		t.Error("Completion command should exist")
	}

	// Test that completion subcommands exist
	bashCmd := findCommand(completionCmd, "bash")
	if bashCmd == nil {
		t.Error("Bash completion command should exist")
	}

	zshCmd := findCommand(completionCmd, "zsh")
	if zshCmd == nil {
		t.Error("Zsh completion command should exist")
	}

	fishCmd := findCommand(completionCmd, "fish")
	if fishCmd == nil {
		t.Error("Fish completion command should exist")
	}

	powershellCmd := findCommand(completionCmd, "powershell")
	if powershellCmd == nil {
		t.Error("Powershell completion command should exist")
	}
}

func TestCompletionCommandStructure(t *testing.T) {
	cli := NewCLI()

	// Test all completion commands have proper structure
	commands := []struct {
		name    string
		builder func() *cobra.Command
	}{
		{"completion", cli.buildCompletionCommand},
		{"bash", cli.buildBashCompletionCommand},
		{"zsh", cli.buildZshCompletionCommand},
		{"fish", cli.buildFishCompletionCommand},
		{"powershell", cli.buildPowershellCompletionCommand},
	}

	for _, tt := range commands {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.builder()

			if cmd == nil {
				t.Errorf("%s completion command should not be nil", tt.name)
				return
			}

			if cmd.Short == "" {
				t.Errorf("%s completion command should have short description", tt.name)
			}

			if cmd.RunE == nil && tt.name != "completion" {
				t.Errorf("%s completion command should have RunE function", tt.name)
			}
		})
	}
}
