package cmd

import (
	"github.com/lorislab/changelog/changelog"
	"github.com/lorislab/changelog/github"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type generateFlags struct {
	Owner   string `mapstructure:"owner"`
	Repo    string `mapstructure:"repo"`
	Token   string `mapstructure:"token"`
	Version string `mapstructure:"version"`
	File    string `mapstructure:"file"`
}

func init() {
	rootCmd.AddCommand(generateCmd)
	addFlagR(generateCmd, "owner", "", "", "project owner")
	addFlagR(generateCmd, "repo", "", "", "repository name")
	addFlag(generateCmd, "token", "", "", "access token")
	addFlagR(generateCmd, "version", "", "", "release version")
	addFlag(generateCmd, "file", "f", "changelog.yaml", "changelog definition")

}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate changelog",
	Long:  `Generate change for the release`,
	Run: func(cmd *cobra.Command, args []string) {

		options := readGenerateFlags()

		// current implementation support only github
		client := github.CreateClient(options.Owner, options.Repo, options.Token)

		// Groups: []*changelog.Group{
		// 	{Title: "Major changes", Labels: []string{"release/super-fearure"}, Items: []changelog.Item{}},
		// 	{Title: "Complete changelog", Labels: []string{"bug", "enhancement"}, Items: []changelog.Item{}},
		// },
		// create changelog base on the configuration
		changelog := changelog.Changelog{
			Version: options.Version,
			File:    options.File,
			Client:  client,
		}

		// Initialize changelog
		changelog.Init()

		// find all issue for the version
		changelog.FindVersionIssues()

		// generate release body
		output := changelog.GenerateBody()
		log.Debugf("\n%s", output)

		// changelog.CreateRelease()
	},
}

func readGenerateFlags() generateFlags {
	options := generateFlags{}
	err := viper.Unmarshal(&options)
	if err != nil {
		panic(err)
	}
	log.Debug(options)
	return options
}

func addFlag(command *cobra.Command, name, shorthand, value, usage string) *pflag.Flag {
	return addFlagExt(command, name, shorthand, value, usage, false)
}

func addFlagR(command *cobra.Command, name, shorthand, value, usage string) *pflag.Flag {
	return addFlagExt(command, name, shorthand, value, usage+" (mandatory)", true)
}

func addFlagExt(command *cobra.Command, name, shorthand, value, usage string, required bool) *pflag.Flag {
	command.Flags().StringP(name, shorthand, value, usage)
	if required {
		command.MarkFlagRequired(name)
	}
	return addViper(command, name)
}

func addStringSliceFlag(command *cobra.Command, name, shorthand string, value []string, usage string) *pflag.Flag {
	command.Flags().StringSliceP(name, shorthand, value, usage)
	return addViper(command, name)
}

func addViper(command *cobra.Command, name string) *pflag.Flag {
	f := command.Flags().Lookup(name)
	err := viper.BindPFlag(name, f)
	if err != nil {
		panic(err)
	}
	return f
}
