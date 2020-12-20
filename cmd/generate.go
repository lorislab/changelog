package cmd

import (
	"fmt"

	"github.com/lorislab/changelog/changelog"
	"github.com/lorislab/changelog/github"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(generateCmd)
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate changelog",
	Long:  `Generate change for the release`,
	Run: func(cmd *cobra.Command, args []string) {

		// current implementation support only github
		client := github.CreateClient("andrejpetras", "release-notes", "89c5625dbf12383d6a9c1bcb6ffa457b93d05d91")

		// create changelog base on the configuration
		changelog := changelog.Changelog{
			Version:     "2.0.0",
			Description: "Description",
			Groups: []*changelog.Group{
				{Title: "Major changes", Labels: []string{"release/super-fearure"}, Items: []changelog.Item{}},
				{Title: "Complete changelog", Labels: []string{"bug", "enhancement"}, Items: []changelog.Item{}},
			},
			Client: client,
		}

		// check the version
		changelog.CheckSemVer()

		// find all issue for the version
		changelog.FindVersionIssues()

		// generate release body
		output := changelog.GenerateBody()
		fmt.Println(output)

		//changelog.CreateRelease()
	},
}
