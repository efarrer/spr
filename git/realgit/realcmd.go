package realgit

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/ejoffe/spr/config"
	gogit "github.com/go-git/go-git/v5"
	gogitconfig "github.com/go-git/go-git/v5/config"
	gogitplumbing "github.com/go-git/go-git/v5/plumbing"
	"github.com/rs/zerolog/log"
)

// NewGitCmd returns a new git cmd instance
func NewGitCmd(cfg *config.Config) *gitcmd {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	repo, err := gogit.PlainOpenWithOptions(cwd, &gogit.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		fmt.Printf("%s is not a git repository\n", cwd)
		os.Exit(2)
	}

	wt, err := repo.Worktree()
	if err != nil {
		fmt.Printf("%s is a bare git repository\n", cwd)
		os.Exit(2)
	}

	rootdir := wt.Filesystem.Root()
	rootdir = strings.TrimSpace(maybeAdjustPathPerPlatform(rootdir))

	return &gitcmd{
		config:  cfg,
		repo:    repo,
		rootdir: rootdir,
		stderr:  os.Stderr,
	}
}

func maybeAdjustPathPerPlatform(rawRootDir string) string {
	if strings.HasPrefix(rawRootDir, "/cygdrive") {
		// This is safe to run also on "proper" Windows paths
		cmd := exec.Command("cygpath", []string{"-w", rawRootDir}...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			panic(err)
		}
		return string(out)
	}

	return rawRootDir
}

type gitcmd struct {
	config  *config.Config
	repo    *gogit.Repository
	rootdir string
	stderr  io.Writer
}

func (c *gitcmd) Git(argStr string, output *string) error {
	return c.GitWithEditor(argStr, output, "/usr/bin/true")
}

func (c *gitcmd) MustGit(argStr string, output *string) {
	err := c.Git(argStr, output)
	if err != nil {
		panic(err)
	}
}

func (c *gitcmd) GitWithEditor(argStr string, output *string, editorCmd string) error {
	// runs a git command
	//  if output is not nil it will be set to the output of the command

	// Rebase disabled
	_, noRebaseFlag := os.LookupEnv("SPR_NOREBASE")
	if (c.config.User.NoRebase || noRebaseFlag) && strings.HasPrefix(argStr, "rebase") {
		return nil
	}

	log.Debug().Msg("git " + argStr)
	if c.config.User.LogGitCommands {
		fmt.Printf("> git %s\n", argStr)
	}
	args := []string{
		"-c", fmt.Sprintf("core.editor=%s", editorCmd),
		"-c", "commit.verbose=false",
		"-c", "rebase.abbreviateCommands=false",
		"-c", fmt.Sprintf("sequence.editor=%s", editorCmd),
	}
	args = append(args, strings.Split(argStr, " ")...)
	cmd := exec.Command("git", args...)
	cmd.Dir = c.rootdir

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)

		if parts[1] != "" && strings.ToUpper(parts[0]) != "EDITOR" {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", parts[0], parts[1]))
		}
	}

	if output != nil {
		out, err := cmd.CombinedOutput()
		*output = strings.TrimSpace(string(out))
		if err != nil {
			fmt.Fprintf(c.stderr, "git error: %s", string(out))
			return err
		}
	} else {
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Fprintf(c.stderr, "git error: %s", string(out))
			return err
		}
	}
	return nil
}

func (c *gitcmd) RootDir() string {
	return c.rootdir
}

func (c *gitcmd) SetRootDir(newroot string) {
	c.rootdir = newroot
}

func (c *gitcmd) SetStderr(stderr io.Writer) {
	c.stderr = stderr
}

func (c *gitcmd) DeleteRemoteBranch(ctx context.Context, branch string) error {
	remoteName := c.config.Repo.GitHubRemote

	remote, err := c.repo.Remote(remoteName)
	if err != nil {
		return fmt.Errorf("getting remote %s %w", remoteName, err)
	}

	// Construct the reference name for branch
	refName := gogitplumbing.NewBranchReferenceName(branch)

	pushOptions := gogit.PushOptions{
		RemoteName: remoteName,
		// Nothing before the colon says to push nothing to the destination branch (which deletes it).
		RefSpecs: []gogitconfig.RefSpec{gogitconfig.RefSpec(fmt.Sprintf(":%s", refName))},
	}

	// Delete the remote branch
	err = remote.Push(&pushOptions)
	if err != nil {
		return fmt.Errorf("removing remote branch %s %w", branch, err)
	}

	return nil
}

// GetLocalBranchShortName returns the local branch short name (like "main")
func (c *gitcmd) GetLocalBranchShortName() (string, error) {
	ref, err := c.repo.Head()
	if err != nil {
		return "", fmt.Errorf("getting HEAD %w", err)
	}

	return string(ref.Name().Short()), nil
}
