package main

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	log "github.com/gothew/l-og"
	"github.com/moby/term"
	// "github.com/docker/docker/pkg/stdcopy"
)

func listImage(ctx context.Context, cli *client.Client) ([]image.Summary, error) {
	return cli.ImageList(ctx, image.ListOptions{})
}

func pullImage(ctx context.Context, cli *client.Client) (io.ReadCloser, error) {
	reader, err := cli.ImagePull(ctx, "archlinux", image.PullOptions{
		Platform: "linux/amd64",
	})
	if err != nil {
		return nil, err
	}

	defer reader.Close()
	// cli.ImagePull is asynchronous.
	// The reader needs to be read completely for the pull operation to complete.
	// If stdout is not required, consider using io.Discard instead of os.Stdout.
	io.Copy(os.Stdout, reader)
	return reader, nil
}

func createContainer(ctx context.Context, cli *client.Client) (io.ReadCloser, error) {
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "alpine",
		Cmd:   []string{"echo", "hello world"},
		Tty:   false,
	}, nil, nil, nil, "")
	if err != nil {
		return nil, err
	}
	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return nil, err
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}
	return out, nil
}

func createContainerTTY(ctx context.Context, cli *client.Client, config *container.Config, hostConfig *container.HostConfig) {
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "")

	if resp.ID == "" {
		if _, err := pullImage(ctx, cli); err != nil {
			log.Error(err)
			panic(err)
		}
	}

	if err != nil {
		log.Error(err)
		panic(err)
	}

	// attach terminal
	hijack, err := cli.ContainerAttach(ctx, resp.ID, container.AttachOptions{
		Stream: true, Stdin: true, Stdout: true, Stderr: true, Logs: true,
	})

	if err != nil {
		panic(err)
	}

	defer hijack.Close()

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		panic(err)
	}

	inFd, _ := term.GetFdInfo(os.Stdin)
	state, err := term.SetRawTerminal(inFd)
	if err != nil {
		panic(err)
	}

	defer term.RestoreTerminal(inFd, state)

	go func() {
		_, _ = io.Copy(hijack.Conn, os.Stdin)
	}()

	_, _ = io.Copy(os.Stdout, hijack.Conn)
}

func buildDockerfile(ctx context.Context, dockerfile string, cli *client.Client) string {
	imageTag := "lab:latest"
	tarBuf := new(bytes.Buffer)
	tw := tar.NewWriter(tarBuf)

	file, err := os.Open(dockerfile)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	stat, _ := file.Stat()
	hdr := &tar.Header{
		Name: filepath.Base(dockerfile),
		Mode: 0600,
		Size: stat.Size(),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		panic(err)
	}

	if _, err := io.Copy(tw, file); err != nil {
		panic(err)
	}
	tw.Close()

	buildResp, err := cli.ImageBuild(ctx, tarBuf, build.ImageBuildOptions{
		Tags:       []string{imageTag},
		Dockerfile: filepath.Base(dockerfile),
		Remove:     true,
	})

	if err != nil {
		panic(err)
	}
	defer buildResp.Body.Close()
	_, _ = io.Copy(os.Stdout, buildResp.Body)
	log.Info("Build complete")
	return imageTag
}

func main() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	dockerfile := "dockers/Dockerfile.work"
	imageTag := buildDockerfile(ctx, dockerfile, cli)

	abs, _ := filepath.Abs("../saas-infra")
	_, err = os.Stat(abs)
	if os.IsNotExist(err) {
		log.Fatalf("La ruta no existe: %s", abs)
	}
	createContainerTTY(ctx, cli, &container.Config{
		Image:        imageTag,
		Cmd:          []string{"/bin/sh"},
		Tty:          true,
		OpenStdin:    true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: abs,
				Target: "/lab",
			},
		},
	})
	// stdcopy.StdCopy(os.Stdout, os.Stderr, out)
}
