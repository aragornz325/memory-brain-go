package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	WorkspaceFlag string
	ProjectFlag   string
	JsonFlag      bool
)

var RootCmd = &cobra.Command{
	Use:   "memory",
	Short: "Memory Brain CLI allows you to remember and search developer context.",
	Long: `Memory Brain CLI connects to a Memory Brain API server to record
and query semantically indexed developers' memories and logs.`,
}

func Execute() {
	RootCmd.PersistentFlags().StringVar(&WorkspaceFlag, "workspace", "", "Workspace slug (or MEMORY_BRAIN_WORKSPACE env)")
	RootCmd.PersistentFlags().StringVar(&ProjectFlag, "project", "", "Project slug (or MEMORY_BRAIN_PROJECT env)")
	RootCmd.PersistentFlags().BoolVar(&JsonFlag, "json", false, "Output results in raw JSON format")

	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
