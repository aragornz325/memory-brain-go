package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	ProjectSlugFlag      string
	ProjectWorkspaceFlag string
)

func init() {
	projectCmd.AddCommand(projectCreateCmd)
	projectCreateCmd.Flags().StringVar(&ProjectSlugFlag, "slug", "", "Unique URL-friendly slug for the project (required)")
	projectCreateCmd.Flags().StringVar(&ProjectWorkspaceFlag, "workspace", "", "Slug of the parent workspace (required)")
	_ = projectCreateCmd.MarkFlagRequired("slug")
	_ = projectCreateCmd.MarkFlagRequired("workspace")

	RootCmd.AddCommand(projectCmd)
}

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
}

var projectCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new project in a workspace",
	Run: func(cmd *cobra.Command, args []string) {
		config := LoadCliConfig(WorkspaceFlag, ProjectFlag)
		client := NewClient(config.BaseURL, config.APIKey)

		payload := &CreateProjectRequest{
			WorkspaceSlug: ProjectWorkspaceFlag,
			Slug:          ProjectSlugFlag,
		}

		result, err := client.CreateProject(context.Background(), payload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Project creation failed: %v\n", err)
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
