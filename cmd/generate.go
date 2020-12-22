package cmd

import (
	changelog "github.com/lorislab/changelog/api"
	"github.com/lorislab/changelog/github"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type generateFlags struct {
	Repository   string `mapstructure:"repository"`
	Token        string `mapstructure:"token"`
	Version      string `mapstructure:"version"`
	File         string `mapstructure:"file"`
	Release      bool   `mapstructure:"create-release"`
	CloseVersion bool   `mapstructure:"close-version"`
	Output       bool   `mapstructure:"console"`
}

func (f generateFlags) log() log.Fields {
	return log.Fields{
		"repository": f.Repository,
		"version":    f.Version,
		"file":       f.File,
		"release":    f.Release,
		"close":      f.CloseVersion,
		"output":     f.Output,
	}
}

func init() {
	rootCmd.AddCommand(generateCmd)
	addFlag(generateCmd, "repository", "r", "", "repository name")
	addFlag(generateCmd, "token", "t", "", "access token")
	addFlag(generateCmd, "version", "e", "", "release version")
	addBoolFlag(generateCmd, "create-release", "", false, "create release and changelog")
	addBoolFlag(generateCmd, "close-version", "", false, "close version")
	addBoolFlag(generateCmd, "console", "", false, "write changelog to the console")
	addFlag(generateCmd, "file", "f", "changelog.yaml", "changelog definition")

}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate changelog",
	Long:  `Generate change for the release`,
	Run: func(cmd *cobra.Command, args []string) {

		options := readGenerateFlags()

		// current implementation support only github
		client, version := github.Init(options.Repository, options.Token)

		// replace version with a input parameter
		if len(options.Version) > 0 {
			version = options.Version
		}

		// check version
		if len(version) == 0 {
			log.WithFields(log.Fields{
				"version":         version,
				"options-version": options.Version,
			}).Fatal("Version is empty!")
		}

		// create changelog base on the configuration
		changelog := changelog.Changelog{
			Version: version,
			File:    options.File,
			Client:  client,
		}

		// Initialize changelog
		changelog.Init()

		// find all issue for the version
		changelog.FindVersionIssues()

		// generate release body
		changelog.GenerateBody()
		if options.Output {
			log.Infof("\n%s", changelog.Body)
		}

		// create release
		if options.Release {
			changelog.CreateRelease()
		}

		// close version
		if options.CloseVersion {
			changelog.CloseVersion()
		}
	},
}

func readGenerateFlags() generateFlags {
	options := generateFlags{}
	err := viper.Unmarshal(&options)
	if err != nil {
		log.Fatal(err)
	}

	log.WithFields(options.log()).Debug("Load configuration")
	return options
}

func addFlag(command *cobra.Command, name, shorthand, value, usage string) *pflag.Flag {
	return addFlagExt(command, name, shorthand, value, usage, false)
}

func addFlagExt(command *cobra.Command, name, shorthand, value, usage string, required bool) *pflag.Flag {
	command.Flags().StringP(name, shorthand, value, usage)
	if required {
		err := command.MarkFlagRequired(name)
		if err != nil {
			log.Panic(err)
		}
	}
	return addViper(command, name)
}

func addBoolFlag(command *cobra.Command, name, shorthand string, value bool, usage string) *pflag.Flag {
	command.Flags().BoolP(name, shorthand, value, usage)
	return addViper(command, name)
}

func addViper(command *cobra.Command, name string) *pflag.Flag {
	f := command.Flags().Lookup(name)
	err := viper.BindPFlag(name, f)
	if err != nil {
		log.Panic(err)
	}
	return f
}
