package pulumi

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/pkg/v2/backend"
	"github.com/pulumi/pulumi/pkg/v2/engine"
	"github.com/pulumi/pulumi/pkg/v2/resource/deploy"
	"github.com/pulumi/pulumi/pkg/v2/util/cancel"
	"github.com/pulumi/pulumi/sdk/v2/go/common/diag"
	"github.com/pulumi/pulumi/sdk/v2/go/common/diag/colors"
	"github.com/pulumi/pulumi/sdk/v2/go/common/util/ciutil"
	"github.com/pulumi/pulumi/sdk/v2/go/common/util/cmdutil"
	"github.com/pulumi/pulumi/sdk/v2/go/common/util/contract"
	"github.com/pulumi/pulumi/sdk/v2/go/common/util/gitutil"
	"github.com/pulumi/pulumi/sdk/v2/go/common/util/logging"
	"github.com/pulumi/pulumi/sdk/v2/go/common/util/result"
	"gopkg.in/src-d/go-git.v4"
)

type cancellationScope struct {
	context *cancel.Context
	sigint  chan os.Signal
	done    chan bool
}

func (s *cancellationScope) Context() *cancel.Context {
	return s.context
}

func (s *cancellationScope) Close() {
	signal.Stop(s.sigint)
	close(s.sigint)
	<-s.done
}

type cancellationScopeSource int

var cancellationScopes = backend.CancellationScopeSource(cancellationScopeSource(0))

func getUpdateMetadata(msg, root string) (*backend.UpdateMetadata, error) {
	m := &backend.UpdateMetadata{
		Message:     msg,
		Environment: make(map[string]string),
	}

	if err := addGitMetadata(root, m); err != nil {
		logging.V(3).Infof("errors detecting git metadata: %s", err)
	}

	addCIMetadataToEnvironment(m.Environment)

	return m, nil
}

// addGitMetadata populate's the environment metadata bag with Git-related values.
func addGitMetadata(repoRoot string, m *backend.UpdateMetadata) error {
	var allErrors *multierror.Error

	// Gather git-related data as appropriate. (Returns nil, nil if no repo found.)
	repo, err := gitutil.GetGitRepository(repoRoot)
	if err != nil {
		return errors.Wrapf(err, "detecting Git repository")
	}
	if repo == nil {
		return nil
	}

	if err := AddGitRemoteMetadataToMap(repo, m.Environment); err != nil {
		allErrors = multierror.Append(allErrors, err)
	}

	if err := addGitCommitMetadata(repo, repoRoot, m); err != nil {
		allErrors = multierror.Append(allErrors, err)
	}

	return allErrors.ErrorOrNil()
}

// AddGitRemoteMetadataToMap reads the given git repo and adds its metadata to the given map bag.
func AddGitRemoteMetadataToMap(repo *git.Repository, env map[string]string) error {
	var allErrors *multierror.Error

	// Get the remote URL for this repo.
	remoteURL, err := gitutil.GetGitRemoteURL(repo, "origin")
	if err != nil {
		return errors.Wrap(err, "detecting Git remote URL")
	}
	if remoteURL == "" {
		return nil
	}

	// Check if the remote URL is a GitHub or a GitLab URL.
	if err := addVCSMetadataToEnvironment(remoteURL, env); err != nil {
		allErrors = multierror.Append(allErrors, err)
	}

	return allErrors.ErrorOrNil()
}

func addVCSMetadataToEnvironment(remoteURL string, env map[string]string) error {
	// GitLab, Bitbucket, Azure DevOps etc. repo slug if applicable.
	// We don't require a cloud-hosted VCS, so swallow errors.
	vcsInfo, err := gitutil.TryGetVCSInfo(remoteURL)
	if err != nil {
		return errors.Wrap(err, "detecting VCS project information")
	}
	env[backend.VCSRepoOwner] = vcsInfo.Owner
	env[backend.VCSRepoName] = vcsInfo.Repo
	env[backend.VCSRepoKind] = vcsInfo.Kind

	return nil
}

func addGitCommitMetadata(repo *git.Repository, repoRoot string, m *backend.UpdateMetadata) error {
	// When running in a CI/CD environment, the current git repo may be running from a
	// detached HEAD and may not have have the latest commit message. We fall back to
	// CI-system specific environment variables when possible.
	ciVars := ciutil.DetectVars()

	// Commit at HEAD
	head, err := repo.Head()
	if err != nil {
		return errors.Wrap(err, "getting repository HEAD")
	}

	hash := head.Hash()
	m.Environment[backend.GitHead] = hash.String()
	commit, commitErr := repo.CommitObject(hash)
	if commitErr != nil {
		return errors.Wrap(commitErr, "getting HEAD commit info")
	}

	// If in detached head, will be "HEAD", and fallback to use value from CI/CD system if possible.
	// Otherwise, the value will be like "refs/heads/master".
	headName := head.Name().String()
	if headName == "HEAD" && ciVars.BranchName != "" {
		headName = ciVars.BranchName
	}
	if headName != "HEAD" {
		m.Environment[backend.GitHeadName] = headName
	}

	// If there is no message set manually, default to the Git commit's title.
	msg := strings.TrimSpace(commit.Message)
	if msg == "" && ciVars.CommitMessage != "" {
		msg = ciVars.CommitMessage
	}
	if m.Message == "" {
		m.Message = gitCommitTitle(msg)
	}

	// Store committer and author information.
	m.Environment[backend.GitCommitter] = commit.Committer.Name
	m.Environment[backend.GitCommitterEmail] = commit.Committer.Email
	m.Environment[backend.GitAuthor] = commit.Author.Name
	m.Environment[backend.GitAuthorEmail] = commit.Author.Email

	// If the worktree is dirty, set a bit, as this could be a mistake.
	isDirty, err := isGitWorkTreeDirty(repoRoot)
	if err != nil {
		return errors.Wrapf(err, "checking git worktree dirty state")
	}
	m.Environment[backend.GitDirty] = strconv.FormatBool(isDirty)

	return nil
}

// gitCommitTitle turns a commit message into its title, simply by taking the first line.
func gitCommitTitle(s string) string {
	if ixCR := strings.Index(s, "\r"); ixCR != -1 {
		s = s[:ixCR]
	}
	if ixLF := strings.Index(s, "\n"); ixLF != -1 {
		s = s[:ixLF]
	}
	return s
}

// addCIMetadataToEnvironment populates the environment metadata bag with CI/CD-related values.
func addCIMetadataToEnvironment(env map[string]string) {
	// Add the key/value pair to env, if there actually is a value.
	addIfSet := func(key, val string) {
		if val != "" {
			env[key] = val
		}
	}

	// Use our built-in CI/CD detection logic.
	vars := ciutil.DetectVars()
	if vars.Name == "" {
		return
	}
	env[backend.CISystem] = string(vars.Name)
	addIfSet(backend.CIBuildID, vars.BuildID)
	addIfSet(backend.CIBuildNumer, vars.BuildNumber)
	addIfSet(backend.CIBuildType, vars.BuildType)
	addIfSet(backend.CIBuildURL, vars.BuildURL)
	addIfSet(backend.CIPRHeadSHA, vars.SHA)
	addIfSet(backend.CIPRNumber, vars.PRNumber)
}

// anyWriter is an io.Writer that will set itself to `true` iff any call to `anyWriter.Write` is made with a
// non-zero-length slice. This can be used to determine whether or not any data was ever written to the writer.
type anyWriter bool

func (w *anyWriter) Write(d []byte) (int, error) {
	if len(d) > 0 {
		*w = true
	}
	return len(d), nil
}

// isGitWorkTreeDirty returns true if the work tree for the current directory's repository is dirty.
func isGitWorkTreeDirty(repoRoot string) (bool, error) {
	gitBin, err := exec.LookPath("git")
	if err != nil {
		return false, err
	}

	gitStatusCmd := exec.Command(gitBin, "status", "--porcelain", "-z")
	var anyOutput anyWriter
	var stderr bytes.Buffer
	gitStatusCmd.Dir = repoRoot
	gitStatusCmd.Stdout = &anyOutput
	gitStatusCmd.Stderr = &stderr
	if err = gitStatusCmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			ee.Stderr = stderr.Bytes()
		}
		return false, errors.Wrapf(err, "'git status' failed")
	}

	return bool(anyOutput), nil
}

func (cancellationScopeSource) NewScope(events chan<- engine.Event, isPreview bool) backend.CancellationScope {
	cancelContext, cancelSource := cancel.NewContext(context.Background())

	c := &cancellationScope{
		context: cancelContext,
		sigint:  make(chan os.Signal),
		done:    make(chan bool),
	}

	go func() {
		for range c.sigint {
			// If we haven't yet received a SIGINT, call the cancellation func. Otherwise call the termination
			// func.
			if cancelContext.CancelErr() == nil {
				message := "^C received; cancelling. If you would like to terminate immediately, press ^C again.\n"
				if !isPreview {
					message += colors.BrightRed + "Note that terminating immediately may lead to orphaned resources " +
						"and other inconsistencies.\n" + colors.Reset
				}
				events <- engine.NewEvent(engine.StdoutColorEvent, engine.StdoutEventPayload{
					Message: message,
					Color:   colors.Always,
				})

				cancelSource.Cancel()
			} else {
				message := colors.BrightRed + "^C received; terminating" + colors.Reset
				events <- engine.NewEvent(engine.StdoutColorEvent, engine.StdoutEventPayload{
					Message: message,
					Color:   colors.Always,
				})

				cancelSource.Terminate()
			}
		}
		close(c.done)
	}()
	signal.Notify(c.sigint, os.Interrupt)

	return c
}

// PrintEngineResult optionally provides a place for the CLI to provide human-friendly error
// messages for messages that can happen during normal engine operation.
func PrintEngineResult(res result.Result) result.Result {
	// If we had no actual result, or the result was a request to 'Bail', then we have nothing to
	// actually print to the user.
	if res == nil || res.IsBail() {
		return res
	}

	err := res.Error()

	switch e := err.(type) {
	case deploy.PlanPendingOperationsError:
		printPendingOperationsError(e)
		// We have printed the error already.  Should just bail at this point.
		return result.Bail()
	case engine.DecryptError:
		printDecryptError(e)
		// We have printed the error already.  Should just bail at this point.
		return result.Bail()
	default:
		// Caller will handle printing of this true error in a generalized fashion.
		return res
	}
}

func printPendingOperationsError(e deploy.PlanPendingOperationsError) {
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	fprintf(writer,
		"the current deployment has %d resource(s) with pending operations:\n", len(e.Operations))

	for _, op := range e.Operations {
		fprintf(writer, "  * %s, interrupted while %s\n", op.Resource.URN, op.Type)
	}

	fprintf(writer, `
These resources are in an unknown state because the Pulumi CLI was interrupted while
waiting for changes to these resources to complete. You should confirm whether or not the
operations listed completed successfully by checking the state of the appropriate provider.
For example, if you are using AWS, you can confirm using the AWS Console.

Once you have confirmed the status of the interrupted operations, you can repair your stack
using 'pulumi stack export' to export your stack to a file. For each operation that succeeded,
remove that operation from the "pending_operations" section of the file. Once this is complete,
use 'pulumi stack import' to import the repaired stack.

refusing to proceed`)
	contract.IgnoreError(writer.Flush())

	cmdutil.Diag().Errorf(diag.RawMessage("" /*urn*/, buf.String()))
}

func printDecryptError(e engine.DecryptError) {
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	fprintf(writer, "failed to decrypt encrypted configuration value '%s': %s", e.Key, e.Err.Error())
	fprintf(writer, `
This can occur when a secret is copied from one stack to another. Encryption of secrets is done per-stack and
it is not possible to share an encrypted configuration value across stacks.

You can re-encrypt your configuration by running 'pulumi config set %s [value] --secret' with your
new stack selected.

refusing to proceed`, e.Key)
	contract.IgnoreError(writer.Flush())
	cmdutil.Diag().Errorf(diag.RawMessage("" /*urn*/, buf.String()))
}

// Quick and dirty utility function for printing to writers that we know will never fail.
func fprintf(writer io.Writer, msg string, args ...interface{}) {
	_, err := fmt.Fprintf(writer, msg, args...)
	contract.IgnoreError(err)
}
