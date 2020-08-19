package cmd

import (
	"fmt"

	do "github.com/pulumi/pulumi-digitalocean/sdk/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cpulumi "github.com/darkowlzz/clouddev/pulumi"
)

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Provision cloud environment",
	Long:  `Provision cloud environment as per the provided configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("up called")

		stackName := viper.GetString("pulumi.stack")
		fmt.Println("StackName:", stackName)

		// TODO: Check if it already exists before initialization.
		if err := cpulumi.InitializeStack(stackName); err != nil {
			return err
		}

		if err := up(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(upCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// upCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// upCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func up() error {
	dropletName := viper.GetString("name")

	// Create droplet args.
	droplet := &do.DropletArgs{
		Image:   pulumi.String(viper.GetString("image")),
		Region:  pulumi.String(viper.GetString("region")),
		Size:    pulumi.String(viper.GetString("size")),
		SshKeys: pulumi.StringArray{},
		Tags:    pulumi.StringArray{},
	}

	// Read ssh key IDs from the config file.
	sshkeys := pulumi.StringArray{}
	keyIDs := viper.GetStringSlice("sshKeys")
	for _, id := range keyIDs {
		sshkeys = append(sshkeys, pulumi.String(id))
	}
	droplet.SshKeys = sshkeys

	// Read tags from the config file.
	tags := pulumi.StringArray{}
	readTags := viper.GetStringSlice("tags")
	for _, tag := range readTags {
		tags = append(tags, pulumi.String(tag))
	}
	droplet.Tags = tags

	pulumi.Run(func(ctx *pulumi.Context) error {
		// Image:  pulumi.String("ubuntu-18-04-x64"),
		droplet, err := do.NewDroplet(ctx, dropletName, droplet)
		if err != nil {
			return err
		}

		ctx.Export("name", droplet.Name)
		ctx.Export("ip", droplet.Ipv4Address)

		return nil
	})

	return nil
}
