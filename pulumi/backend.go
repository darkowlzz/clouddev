package pulumi

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/pkg/v2/backend/filestate"
	"github.com/pulumi/pulumi/sdk/v2/go/common/util/cmdutil"
)

func SetupBackend() error {
	cloudURL := filestate.FilePathPrefix + "~"

	be, err := filestate.Login(cmdutil.Diag(), cloudURL)
	if err != nil {
		return errors.Wrap(err, "problem logging in")
	}

	fmt.Printf("Logged in to %s (%s)\n", be.Name(), be.URL())

	return nil
}
