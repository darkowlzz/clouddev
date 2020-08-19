package pulumi

import (
	"context"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/pkg/v2/backend"
	"github.com/pulumi/pulumi/pkg/v2/backend/filestate"
	"github.com/pulumi/pulumi/pkg/v2/backend/state"
	"github.com/pulumi/pulumi/sdk/v2/go/common/util/cmdutil"
	"github.com/pulumi/pulumi/sdk/v2/go/common/workspace"
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

	// Create new stack.
	ctx := context.Background()
	var createOpts interface{}
	stack, err := b.CreateStack(ctx, stackRef, createOpts)
	if err != nil {
		return err
	}

	// Set current stack.
	if err := state.SetCurrentStack(stack.Ref().String()); err != nil {
		return err
	}

	return nil
}
