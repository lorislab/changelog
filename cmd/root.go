package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	goVersion "go.hein.dev/go-version"
)

var (
	// Used for flags.
	shortened = false
	version   = "dev"
	commit    = "none"
	date      = "unknown"
	output    = "json"
	cfgFile   string
	v         string
	rootCmd   = &cobra.Command{
		Use:   "changelog",
		Short: "Changelog generator for the release",
		Long:  `Generate changelog for the project and create release.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := setUpLogs(os.Stdout, v); err != nil {
				return err
			}
			return nil
		},
	}
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Version will output the current build information",
		Long:  ``,
		Run: func(_ *cobra.Command, _ []string) {
			resp := goVersion.FuncWithOutput(shortened, version, commit, date, output)
			fmt.Print(resp)
		},
	}
)

// Execute executes the root command.
func Execute() {
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
	})
	err := rootCmd.Execute()
	if err != nil {
		log.Panic(err)
	}
}

func init() {
	versionCmd.Flags().BoolVarP(&shortened, "short", "s", false, "Print just the version number.")
	versionCmd.Flags().StringVarP(&output, "output", "o", "json", "Output format. One of 'yaml' or 'json'.")
	rootCmd.AddCommand(versionCmd)

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.changelog.yaml)")
	rootCmd.PersistentFlags().StringVarP(&v, "verbosity", "v", log.WarnLevel.String(), "Log level (debug, info, warn, error, fatal, panic")
}

func er(msg interface{}) {
	fmt.Println("Error:", msg)
	os.Exit(1)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			er(err)
		}

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".changelog")
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetEnvPrefix("CHANGELOG")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		log.Debugf("Using config file: %s", viper.ConfigFileUsed())
	}
}

func setUpLogs(out io.Writer, level string) error {
	log.SetOutput(out)
	lvl, err := log.ParseLevel(level)
	if err != nil {
		return err
	}
	log.SetLevel(lvl)
	return nil
}
