package main

import (
	"context"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/abshkbh/chv-lambda/out/protos"
)

const (
	defaultServerAddress = "localhost:50051"
)

func stopVM(serverAddr string, vmName string) error {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewVMManagementServiceClient(conn)
	ctx := context.Background()

	request := &pb.VMRequest{VmName: vmName}
	_, err = client.StopVM(ctx, request)
	if err != nil {
		return fmt.Errorf("error stopping: %w", err)
	}

	log.Infof("Successfully stopped VM: %s", vmName)
	return nil
}

func destroyVM(serverAddr string, vmName string) error {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewVMManagementServiceClient(conn)
	ctx := context.Background()

	request := &pb.VMRequest{VmName: vmName}
	_, err = client.DestroyVM(ctx, request)
	if err != nil {
		return fmt.Errorf("error destroying: %w", err)
	}

	log.Infof("Successfully destroyed VM: %s", vmName)
	return nil
}

func startVM(serverAddr string, vmName string, langType string) error {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewVMManagementServiceClient(conn)
	ctx := context.Background()

	request := &pb.VMRequest{
		VmName:   vmName,
		LangType: langType,
	}
	_, err = client.StartVM(ctx, request)
	if err != nil {
		return fmt.Errorf("error starting: %w", err)
	}

	log.Infof("Successfully started VM: %s", vmName)
	return nil
}

func main() {
	app := &cli.App{
		Name:  "vm-cli",
		Usage: "A CLI for managing VMs",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "server",
				Aliases: []string{"s"},
				Value:   defaultServerAddress,
				Usage:   "gRPC server address",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start a VM",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Aliases:  []string{"n"},
						Usage:    "Name of the VM to create",
						Required: true,
					},
					// TODO: Deprecate this.
					&cli.StringFlag{
						Name:     "lang-type",
						Aliases:  []string{"lt"},
						Usage:    "If required, the language to support inside the server",
						Required: false,
					},
				},
				Action: func(ctx *cli.Context) error {
					return startVM(ctx.String("server"), ctx.String("name"), ctx.String("lang-type"))
				},
			},
			{
				Name:  "stop",
				Usage: "Stop a VM",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Aliases:  []string{"n"},
						Usage:    "Name of the VM to stop",
						Required: true,
					},
				},
				Action: func(ctx *cli.Context) error {
					return stopVM(ctx.String("server"), ctx.String("name"))
				},
			},
			{
				Name:  "destroy",
				Usage: "Destroy a VM",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Aliases:  []string{"n"},
						Usage:    "Name of the VM to destroy",
						Required: true,
					},
				},
				Action: func(ctx *cli.Context) error {
					return destroyVM(ctx.String("server"), ctx.String("name"))
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
