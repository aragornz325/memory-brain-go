package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	TagsFlag      []string
	SourceFlag    string
	SourceRefFlag string
)

func init() {
	rememberCmd.Flags().StringSliceVar(&TagsFlag, "tags", []string{}, "Comma-separated tags for the memory")
	rememberCmd.Flags().StringVar(&SourceFlag, "source", "cursor", "Source of the memory")
	rememberCmd.Flags().StringVar(&SourceRefFlag, "source-ref", "", "Source reference (e.g. file, line)")
	RootCmd.AddCommand(rememberCmd)
}

var rememberCmd = &cobra.Command{
	Use:   "remember <text>",
	Short: "Store a new memory using free-form text",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "Error: Missing required text argument.")
			_ = cmd.Usage()
			os.Exit(1)
		}

		text := strings.Join(args, " ")
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

		payload := &RememberRequest{
			WorkspaceSlug: config.WorkspaceSlug,
			ProjectSlug:   &config.ProjectSlug,
			Text:          text,
			Tags:          TagsFlag,
			Source:        &SourceFlag,
		}
		if SourceRefFlag != "" {
			payload.SourceRef = &SourceRefFlag
		}

		result, err := client.Remember(context.Background(), payload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Memory Brain request failed: %v\n", err)
			os.Exit(1)
		}

		// Single object outputs are printed as JSON in the NestJS implementation
		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonBytes))
	},
}
