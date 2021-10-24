package provider

import (
	"context"
)

type Provider interface {
	// CreateInstance creates the instance using the provided starter image/ami
	// it should create an instance and make sure that it created successfully
	// TODO: maybe return instance ID?
	CreateInstance(ctx context.Context, runnerInstanceName string) error
	// DestroyInstance destroys the image but does not deregister the runner
	DestroyInstance(ctx context.Context, runnerInstanceName string) error
	// InstanceStatus returns the status of the instance
	InstanceStatus(ctx context.Context, runnerInstanceName string) error

	// WantGithubRegistrationToken tells if provider needs a registration token
	WantGithubRegistrationToken() bool
	WithGithubRegistrationToken(token string)
	WithRunnerType(runnerType string)
}

type providerInstanceStatus int

const (
	providerInstanceStatusInit   providerInstanceStatus = 0
	providerInstanceStatusReady  providerInstanceStatus = 1
	providerInstanceStatusAbsent providerInstanceStatus = 2
)

type BaseProvider struct {
	GithubRegistrationToken string
	RunnerType string
}

// WithRunnerType sets runner type
func (b *BaseProvider)WithRunnerType(runnerType string) {
	b.RunnerType = runnerType
}

// WantGithubRegistrationToken tells if provider needs a registration token
func (b *BaseProvider)WantGithubRegistrationToken() bool { return true }

// WithGithubRegistrationToken sets registration token
func (b *BaseProvider)WithGithubRegistrationToken(token string) {
	b.GithubRegistrationToken = token
}
