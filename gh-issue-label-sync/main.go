package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/google/go-github/v50/github"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

var (
	fromOwner   string
	toOwner     string
	fromRepo    string
	toRepo      string
	accessToken string
)

func init() {
	flag.StringVar(&fromOwner, "from-owner", "", "--from-owner is the owner of the repo you want to sync labels from")
	flag.StringVar(&toOwner, "to-owner", "", "--to-owner is the owner of the repo you want to sync labels to")
	flag.StringVar(&fromRepo, "from-repo", "", "--from-repo is the repo you want to sync labels from")
	flag.StringVar(&toRepo, "to-repo", "", "--to-repo is the repo you want to sync labels to")
	flag.StringVar(&accessToken, "access-token", "", "--access-token is the token to use when calling the GH API")
	flag.Parse()
}

func main() {
	ctx := context.Background()
	logger := lo.Must(zap.NewProduction()).Sugar()
	logger = logger.With("from-repo", fmt.Sprintf("%s/%s", fromOwner, fromRepo), "to-repo", fmt.Sprintf("%s/%s", toOwner, toRepo))
	client := NewClient(ctx, accessToken)
	labels, _ := lo.Must2(client.Issues.ListLabels(ctx, fromOwner, fromRepo, &github.ListOptions{}))

	for _, label := range labels {
		_, _, err := client.Issues.CreateLabel(ctx, toOwner, toRepo, label)
		if err != nil {
			continue
		}
		logger.With("label", label.Name).Infof("added label")
	}
}

func NewClient(ctx context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return client
}
