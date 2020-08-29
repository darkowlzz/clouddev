package pulumi

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	do "github.com/pulumi/pulumi-digitalocean/sdk/go/digitalocean"
	"github.com/pulumi/pulumi/pkg/v2/backend"
	"github.com/pulumi/pulumi/pkg/v2/backend/display"
	"github.com/pulumi/pulumi/pkg/v2/engine"
	"github.com/pulumi/pulumi/sdk/v2/go/common/diag/colors"
	"github.com/pulumi/pulumi/sdk/v2/go/common/workspace"
	// "github.com/pulumi/pulumi/sdk/v2/go/pulumi"
)

func readProject() (*workspace.Project, string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, "", err
	}

	// Now that we got here, we have a path, so we will try to load it.
	path, err := workspace.DetectProjectPathFrom(pwd)
	if err != nil {
		return nil, "", errors.Wrapf(err, "failed to find current Pulumi project because of "+
			"an error when searching for the Pulumi.yaml file (searching upwards from %s)", pwd)
	} else if path == "" {
		return nil, "", fmt.Errorf(
			"no Pulumi.yaml project file found (searching upwards from %s). If you have not "+
				"created a project yet, use `pulumi new` to do so", pwd)
	}
	proj, err := workspace.LoadProject(path)
	if err != nil {
		return nil, "", errors.Wrapf(err, "failed to load Pulumi project located at %q", path)
	}

	return proj, filepath.Dir(path), nil
}

func Up(stackName string, dropletName string, dropletArgs *do.DropletArgs) error {
	b, err := currentBackend()
	if err != nil {
		return err
	}

	if stackName == "" {
		return errors.New("missing stack name")
	}

	if err := b.ValidateStackName(stackName); err != nil {
		return err
	}

	stackRef, err := b.ParseStackReference(stackName)
	if err != nil {
		return err
	}

	ctx := context.Background()

	stack, err := b.GetStack(ctx, stackRef)
	if err != nil {
		return errors.Wrap(err, "could not get stack")
	}

	fmt.Println("Reading PROJECT...")
	proj, root, err := readProject()
	if err != nil {
		return err
	}

	m, err := getUpdateMetadata("", root)
	if err != nil {
		return errors.Wrap(err, "gathering environment metadata")
	}

	// stack.Update(ctx, backend.UpdateOperation{})

	opts := backend.UpdateOptions{
		AutoApprove: true,
		SkipPreview: false,
	}

	opts.Display = display.Options{
		Color: colors.Never,
	}

	opts.Engine = engine.UpdateOptions{}

	_, res := stack.Update(ctx, backend.UpdateOperation{
		Proj:   proj,
		Root:   root,
		Opts:   opts,
		Scopes: cancellationScopes,
		M:      m,
	})

	fmt.Println("RESULT.Err:", res.Error())
	fmt.Println("Result print:", PrintEngineResult(res))

	// pulumi.Run(func(ctx *pulumi.Context) error {
	//     // Image:  pulumi.String("ubuntu-18-04-x64"),
	//     droplet, err := do.NewDroplet(ctx, dropletName, dropletArgs)
	//     if err != nil {
	//         return err
	//     }

	//     ctx.Export("name", droplet.Name)
	//     ctx.Export("ip", droplet.Ipv4Address)

	//     return nil
	// })
	return nil
}
