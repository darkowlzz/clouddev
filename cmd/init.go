package cmd

import (
	"github.com/spf13/cobra"

	"github.com/darkowlzz/clouddev/pulumi"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize backend",
	Long: `Initialize backend for pulumi at ~/. This is used to create pulumi
stacks for cloud deployments.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := pulumi.SetupBackend(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
