package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func errorHandler(err error) {
	if err != nil {
		panic(err)
	}
}

type Docker struct {
	ctx         context.Context
	cli         *client.Client
	imageName   string
	containerID string
}

func newDocker(imageName string) *Docker {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	errorHandler(err)
	return &Docker{ctx, cli, imageName, ""}
}

func (dc *Docker) pullImage() error {
	_, err := dc.cli.ImagePull(dc.ctx, dc.imageName, types.ImagePullOptions{})
	return err
}

func (dc *Docker) createImageAndStart() error {
	resp, err := dc.cli.ContainerCreate(dc.ctx, &container.Config{
		Image: dc.imageName,
		Cmd:   []string{"tail", "-f", "/dev/null"}, //TEMP tail log to test docker status.
	}, nil, nil, nil, "")
	if err != nil {
		return err
	}
	dc.containerID = resp.ID
	err = dc.cli.ContainerStart(dc.ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (dc *Docker) isImageExist() (bool, error) {
	images, err := dc.cli.ImageList(dc.ctx, types.ImageListOptions{})
	if err != nil {
		return false, err
	}

	for _, image := range images {
		if strings.Split(string(image.RepoTags[0]), ":")[0] == dc.imageName {
			return true, nil
		}
	}
	return true, errors.New("not found")
}

//
func (dc *Docker) containerKill() error {
	return dc.cli.ContainerKill(dc.ctx, dc.containerID, "SIGKILL")
}

//
func (dc *Docker) containerStop(timeout time.Duration) error {
	t := time.Duration(timeout) * time.Second
	return dc.cli.ContainerStop(dc.ctx, dc.containerID, &t)
}

func main() {
	dc := newDocker("alpine")
	exist, err := dc.isImageExist()
	if err != nil {
		if exist {
			err = dc.pullImage()
			errorHandler(err)
		} else {
			errorHandler(err)
		}
	}

	err = dc.createImageAndStart()
	errorHandler(err)

	// Next tasks:
	// add stop container timer and kill container timer
	// use channels...
	killInMin := 0
	if killInMin > 0 {
		fmt.Println("KILLING")
		dc.containerKill()
		errorHandler(err)
	}
}
