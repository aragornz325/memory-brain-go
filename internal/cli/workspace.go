package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	WorkspaceSlugFlag string
	WorkspaceNameFlag string
)

func init() {
	workspaceCmd.AddCommand(workspaceCreateCmd)
	workspaceCreateCmd.Flags().StringVar(&WorkspaceSlugFlag, "slug", "", "Unique URL-friendly slug for the workspace (required)")
	workspaceCreateCmd.Flags().StringVar(&WorkspaceNameFlag, "name", "", "Human-readable name of the workspace")
	_ = workspaceCreateCmd.MarkFlagRequired("slug")

	RootCmd.AddCommand(workspaceCmd)
}

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage workspaces",
}

var workspaceCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new workspace",
	Run: func(cmd *cobra.Command, args []string) {
		config := LoadCliConfig(WorkspaceFlag, ProjectFlag)
		client := NewClient(config.BaseURL, config.APIKey)

		payload := &CreateWorkspaceRequest{
			Slug: WorkspaceSlugFlag,
			Name: WorkspaceNameFlag,
		}

		result, err := client.CreateWorkspace(context.Background(), payload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Workspace creation failed: %v\n", err)
			os.Exit(1)
		}

		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonBytes))
	},
}
