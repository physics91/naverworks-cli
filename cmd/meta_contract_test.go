package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCommandMetaContracts(t *testing.T) {
	if err := validateCommandMetaContracts(rootCmd); err != nil {
		t.Fatal(err)
	}
}

func validateCommandMetaContracts(root *cobra.Command) error {
	if root == nil {
		return fmt.Errorf("root command is nil")
	}

	var problems []string
	if !root.SilenceUsage {
		problems = append(problems, "root command must set SilenceUsage=true")
	}
	if !root.SilenceErrors {
		problems = append(problems, "root command must set SilenceErrors=true")
	}

	var walk func(cmd *cobra.Command, lineage []string)
	walk = func(cmd *cobra.Command, lineage []string) {
		commandPath := strings.Join(append(lineage, cmd.Name()), " ")
		if commandPath == "" {
			commandPath = "<root>"
		}

		if !cmd.Hidden {
			if strings.TrimSpace(cmd.Use) == "" {
				problems = append(problems, fmt.Sprintf("%s: missing Use", commandPath))
			}
			if strings.TrimSpace(cmd.Short) == "" {
				problems = append(problems, fmt.Sprintf("%s: missing Short", commandPath))
			}
		}

		commandName := firstUseToken(cmd)
		if isListStyleCommand(commandName) && !strings.Contains(cmd.Short, "목록") {
			problems = append(problems, fmt.Sprintf("%s: list-style command Short must describe list behavior", commandPath))
		}

		if hasPaginationFlag(cmd, "cursor") || hasPaginationFlag(cmd, "all") {
			if !hasPaginationFlag(cmd, "count") {
				problems = append(problems, fmt.Sprintf("%s: pagination commands using --cursor/--all must also declare --count", commandPath))
			}
		}

		childNames := make(map[string]string)
		for _, child := range cmd.Commands() {
			childName := firstUseToken(child)
			if childName == "" {
				continue
			}
			if previous, exists := childNames[childName]; exists {
				problems = append(problems, fmt.Sprintf("%s: duplicate child command name %q (%s, %s)", commandPath, childName, previous, child.CommandPath()))
				continue
			}
			childNames[childName] = child.CommandPath()
			walk(child, append(lineage, cmd.Name()))
		}
	}

	walk(root, nil)

	if len(problems) == 0 {
		return nil
	}
	return fmt.Errorf("command meta contract failures:\n- %s", strings.Join(problems, "\n- "))
}

func firstUseToken(cmd *cobra.Command) string {
	if cmd == nil {
		return ""
	}
	fields := strings.Fields(cmd.Use)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

func isListStyleCommand(name string) bool {
	return name == "list" || strings.HasPrefix(name, "list-")
}

func hasPaginationFlag(cmd *cobra.Command, name string) bool {
	if cmd == nil {
		return false
	}
	return cmd.Flags().Lookup(name) != nil
}
