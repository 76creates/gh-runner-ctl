package provider

import (
	"bytes"
	"html/template"
)

type CloudInitData struct {
	GithubRepo string
	GithubRunnerName string
	GithubRunnerToken string
	GithubRunnerType string
	GithubRunnerUniqueID string
}

// ParseCloudInit generates cloud-init from template
func ParseCloudInit(cloudInitTemplate string, cloudInitData CloudInitData) (*string, error) {
	t, err := template.New("cloud-init").Parse(cloudInitTemplate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, cloudInitData)
	if err != nil {
		return nil, err
	}

	str := buf.String()
	return &str, nil
}
