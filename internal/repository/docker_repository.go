package repository

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

var cli *client.Client

func init() {
	clientWithOpts, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal("docker in docker 连接失败: %w", err)
	}
	ctxPing, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if _, err = cli.Ping(ctxPing); err != nil {
		log.Fatal("docker daemon unreachable: %w", err)
	}
	cli = clientWithOpts
}

func getClient() *client.Client {
	return cli
}

func ReloadConfig(ctx context.Context, containerName string) error {
	log.Printf("start exec cmd container: %s", containerName)
	execOptions := container.ExecOptions{
		Cmd:          []string{"nginx", "-t"},
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
	}
	execResp, err := getClient().ContainerExecCreate(ctx, containerName, execOptions)
	if err != nil {
		return fmt.Errorf("创建exec失败: %w", err)
	}
	attachOptions := container.ExecAttachOptions{}
	attachResp, err := getClient().ContainerExecAttach(ctx, execResp.ID, attachOptions)
	if err != nil {
		return fmt.Errorf("attach exec失败: %w", err)
	}
	defer attachResp.Close()
	outBytes, err := io.ReadAll(attachResp.Reader)
	if err != nil {
		return fmt.Errorf("exec io readAll失败: %w", err)
	}
	inspect, err := cli.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		return fmt.Errorf("inspect exec失败: %w", err)
	}
	if inspect.ExitCode != 0 {
		return fmt.Errorf("inspect exec code %d", inspect.ExitCode)
	}
	log.Printf("inspect exec output: %s", string(outBytes))
	// SIGHUP 对应nginx reload
	if err := getClient().ContainerKill(ctx, containerName, "SIGHUP"); err != nil {
		return fmt.Errorf("发送SIGHUP信号失败: %w", err)
	}
	log.Printf("send SIGHUP success container=%s", containerName)
	return nil
}
