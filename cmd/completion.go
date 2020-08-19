package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Output shell completion code",
	Long: `Output shell completion code.

Installation instructions:

	$ clouddev completion > ~/.clouddev-completion
	$ source ~/.clouddev-completion`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := rootCmd.GenBashCompletion(os.Stdout); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// completionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// completionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
