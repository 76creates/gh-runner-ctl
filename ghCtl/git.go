package ghCtl

import (
	"context"
	"github.com/76creates/runner-cli/log"
	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

// InitClient fetch the github auth client
// this function does not check the validity of the token
func InitClient(ctx context.Context) *github.Client {
	log.Debug("initializing github client")
	token := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: getToken(ctx)},
	)
	o2Client := oauth2.NewClient(ctx, token)

	return github.NewClient(o2Client)
}

// getClient extract github.Client from the context
func getClient(ctx context.Context) *github.Client {
	return ctx.Value("client").(*github.Client)
}

// getToken extract github action token from the context
func getToken(ctx context.Context) string {
	return ctx.Value("github-token").(string)
}

// GetRepoOwner extract repo owner string from the context
func GetRepoOwner(ctx context.Context) string {
	return ctx.Value("github-repo-owner").(string)
}

// GetRepoName extract repo name string from the context
func GetRepoName(ctx context.Context) string {
	return ctx.Value("github-repo-name").(string)
}

// GetRegistrationToken extract the registration token for the runner
func GetRegistrationToken(ctx context.Context) string {
	return ctx.Value("github-registration-token").(string)
}
