package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var LimitFlag int

func init() {
	searchCmd.Flags().IntVar(&LimitFlag, "limit", 5, "Maximum number of search results to retrieve")
	RootCmd.AddCommand(searchCmd)
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search memories semantically",
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
			Limit:         LimitFlag,
		}

		result, err := client.Search(context.Background(), payload)
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

		if len(result) == 0 {
			fmt.Println("No results.")
			return
		}

		for _, item := range result {
			fmt.Printf("- [%s] %s\n", item.ID, item.Text)
			if len(item.Tags) > 0 {
				fmt.Printf("  tags: %s\n", strings.Join(item.Tags, ", "))
			}
		}
	},
}
