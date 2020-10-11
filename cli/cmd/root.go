package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	homedir "github.com/mitchellh/go-homedir"
)

var cfgFile string
var namespace string
var name string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tobs",
	Short: "A CLI tool for The Observablity Stack",
	Long: `The Observability Stack is a tool that uses TimescaleDB as a 
compressed long-term store for time series metrics from Prometheus. This
application is a CLI tool that allows users to quickly access the different
components of Observability.`,
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error

		name, err = cmd.Flags().GetString("name")
		if err != nil {
			return fmt.Errorf("could not read global flag: %w", err)
		}

		namespace, err = cmd.Flags().GetString("namespace")
		if err != nil {
			return fmt.Errorf("could not read global flag: %w", err)
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tobs.yaml)")
	rootCmd.PersistentFlags().StringP("name", "n", "tobs", "Helm release name")
	rootCmd.PersistentFlags().StringP("namespace", "", "default", "Kubernetes namespace")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".tobs" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".tobs")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
