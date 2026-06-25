package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var ContextLimitFlag int

func init() {
	contextCmd.Flags().IntVar(&ContextLimitFlag, "limit", 10, "Maximum number of memories to include in context")
	RootCmd.AddCommand(contextCmd)
}

var contextCmd = &cobra.Command{
	Use:   "context <query>",
	Short: "Retrieve formatted context and related memories for a query",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "Error: Missing required query argument.")
			_ = cmd.Usage()
			os.Exit(1)
		}

		query := strings.Join(args, " ")
		config := LoadCliConfig(WorkspaceFlag, ProjectFlag)

		if config.WorkspaceSlug == "" {
			fmt.Fprintln(os.Stderr, "Error: Workspace slug is required (flag --workspace or MEMORY_BRAIN_WORKSPACE).")
			os.Exit(1)
		}
		if config.ProjectSlug == "" {
			fmt.Fprintln(os.Stderr, "Error: Project slug is required (flag --project or MEMORY_BRAIN_PROJECT).")
			os.Exit(1)
		}

		client := NewClient(config.BaseURL, config.APIKey)

		payload := &SearchRequest{
			WorkspaceSlug: config.WorkspaceSlug,
			ProjectSlug:   &config.ProjectSlug,
			Query:         query,
			Limit:         ContextLimitFlag,
		}

		result, err := client.Context(context.Background(), payload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Memory Brain request failed: %v\n", err)
			os.Exit(1)
		}

		if JsonFlag {
			jsonBytes, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonBytes))
			return
		}

		if result.Context == "" {
			fmt.Println("No context matches found.")
			return
		}

		fmt.Println(result.Context)
	},
}
