package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ejoffe/spr/config"
	"github.com/ejoffe/spr/config/config_parser"
	"github.com/ejoffe/spr/git/realgit"
	"github.com/ejoffe/spr/github"
	"github.com/ejoffe/spr/github/githubclient"
	"github.com/ejoffe/spr/spr"
	ngit "github.com/go-git/go-git/v5"
	gogithub "github.com/google/go-github/v69/github"
	"github.com/jessevdk/go-flags"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	version = "dev"
	commit  = "dversion"
	date    = "unknown"
)

// command line opts
type opts struct {
	Debug   bool `short:"d" long:"debug" description:"Show verbose debug info."`
	Version bool `short:"v" long:"version" description:"Show version."`
	Update  bool `short:"u" long:"update" description:"Run spr update after amend."`
}

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.With().Caller().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	var opts opts
	_, err := flags.Parse(&opts)
	check(err)

	if opts.Version {
		fmt.Printf("amend version : %s : %s : %s\n", version, date, commit[:8])
		os.Exit(0)
	}

	if opts.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	gitcmd := realgit.NewGitCmd(config.DefaultConfig())

	ctx := context.Background()
	cfg := config_parser.ParseConfig(gitcmd)
	client := githubclient.NewGitHubClient(ctx, cfg)
	gitcmd = realgit.NewGitCmd(cfg)
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	repo, err := ngit.PlainOpen(wd)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	goghclient := gogithub.NewClient(nil).WithAuthToken(github.FindToken(cfg.Repo.GitHubHost))

	sd := spr.NewStackedPR(cfg, client, gitcmd, repo, goghclient)
	sd.AmendCommit(ctx)

	if opts.Update {
		sd.UpdatePullRequests(ctx, nil, nil)
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
