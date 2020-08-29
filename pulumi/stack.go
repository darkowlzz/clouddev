package pulumi

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/pkg/v2/backend"
	"github.com/pulumi/pulumi/pkg/v2/backend/filestate"
	"github.com/pulumi/pulumi/pkg/v2/backend/state"
	"github.com/pulumi/pulumi/sdk/v2/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v2/go/common/util/cmdutil"
	"github.com/pulumi/pulumi/sdk/v2/go/common/workspace"
	"gopkg.in/yaml.v3"
)

func currentBackend() (backend.Backend, error) {
	url, err := workspace.GetCurrentCloudURL()
	if err != nil {
		return nil, errors.Wrap(err, "could not get cloud url")
	}

	if filestate.IsFileStateBackendURL(url) {
		return filestate.New(cmdutil.Diag(), url)
	} else {
		return nil, errors.New("non-filestate backend unsupported")
	}
}

func InitializeStack(name string) error {
	// opts := map[string]interface{}{}
	// opts["binary"] = "do"

	// Ensure Pulumi.yaml exists in the current directory.
	project := workspace.Project{
		Name: tokens.PackageName(name),
		Runtime: workspace.NewProjectRuntimeInfo("go", map[string]interface{}{
			"binary": "do",
		}),
	}

	p, err := yaml.Marshal(&project)
	if err != nil {
		return err
	}
	fmt.Println(string(p))

	f, err := os.Create("Pulumi.yaml")
	if err != nil {
		return err
	}

	_, err = f.Write(p)
	if err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	fmt.Println("Pulumi.yaml created!")

	b, err := currentBackend()
	if err != nil {
		return err
	}

	if name == "" {
		return errors.New("missing stack name")
	}

	if err := b.ValidateStackName(name); err != nil {
		return err
	}

	stackRef, err := b.ParseStackReference(name)
	if err != nil {
		return err
	}

	ctx := context.Background()

	var stack backend.Stack
	// Check if the stack exists.
	stack, err = b.GetStack(ctx, stackRef)
	if err == nil {
		if stack == nil {
			// Not found.
			fmt.Println("Creating new stack", name)
			// Create new stack.
			var createOpts interface{}
			stack, err = b.CreateStack(ctx, stackRef, createOpts)
			if err != nil {
				return err
			}
		} else {
			fmt.Println("Using existing stack", stack.Ref().String())
		}
	} else {
		return err
	}

	fmt.Println("Setting current stack...")
	// Set current stack.
	if err := state.SetCurrentStack(stack.Ref().String()); err != nil {
		return errors.Wrap(err, "could not set current stack")
	}

	return nil
}
