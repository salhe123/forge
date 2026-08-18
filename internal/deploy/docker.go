package deploy

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

type Runner interface {
	Run(ctx context.Context, appName, image string) error
}

type DockerRunner struct{}

func (DockerRunner) Run(ctx context.Context, appName, imageName string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("docker client: %w", err)
	}
	defer cli.Close()

	name := ContainerName(appName)
	_ = cli.ContainerRemove(ctx, name, container.RemoveOptions{Force: true})

	pull, err := cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("image pull: %w", err)
	}
	_, _ = io.Copy(io.Discard, pull)
	_ = pull.Close()

	resp, err := cli.ContainerCreate(ctx,
		&container.Config{Image: imageName},
		&container.HostConfig{
			RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyUnlessStopped},
		},
		nil,
		nil,
		name,
	)
	if err != nil {
		return fmt.Errorf("container create: %w", err)
	}
	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("container start: %w", err)
	}
	return nil
}

var safeName = regexp.MustCompile(`[^a-zA-Z0-9_.-]`)

func ContainerName(appName string) string {
	cleaned := strings.ToLower(strings.TrimSpace(appName))
	cleaned = safeName.ReplaceAllString(cleaned, "-")
	if cleaned == "" {
		cleaned = "app"
	}
	return "forge-" + cleaned
}
