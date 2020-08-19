package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/darkowlzz/clouddev/version"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the clouddev version",
	Long:  `Prints the clouddev version.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("version %s\n", version.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
