package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/abshkbh/chv-lambda/out/protos"
	"github.com/abshkbh/chv-lambda/pkg/config"
)

const (
	defaultServerAddress = config.GrpcServerAddr + ":" + config.GrpcServerPort
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

func destroyAllVMs(serverAddr string) error {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewVMManagementServiceClient(conn)
	ctx := context.Background()
	_, err = client.DestroyAllVMs(ctx, &pb.DestroyAllVMsRequest{})
	if err != nil {
		return fmt.Errorf("error destroying all VMs: %w", err)
	}

	log.Info("Successfully destroyed all VMs")
	return nil
}

func startVM(serverAddr string, vmName string, entryPoint string) error {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewVMManagementServiceClient(conn)
	ctx := context.Background()

	request := &pb.StartVMRequest{
		VmName:     vmName,
		EntryPoint: entryPoint,
	}
	resp, err := client.StartVM(ctx, request)
	if err != nil {
		return fmt.Errorf("error starting: %w", err)
	}

	log.Infof("Successfully started VM: %v", resp.VmInfo)
	return nil
}

func printVMInfo(vm *pb.VMInfo) string {
	return fmt.Sprintf("VM: Name=%s, Status=%s, IP=%s, TapDevice=%s",
		vm.VmName,
		strings.TrimPrefix(vm.GetStatus().String(), "VM_STATUS_"),
		vm.Ip,
		vm.TapDeviceName,
	)
}

func listAllVMs(serverAddr string) error {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewVMManagementServiceClient(conn)
	ctx := context.Background()
	resp, err := client.ListAllVMs(ctx, &pb.ListAllVMsRequest{})
	if err != nil {
		return fmt.Errorf("error listing all VMs: %w", err)
	}

	for _, vm := range resp.Vms {
		fmt.Println(printVMInfo(vm))
	}
	return nil
}

func listVM(serverAddr string, vmName string) error {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	client := pb.NewVMManagementServiceClient(conn)
	ctx := context.Background()

	request := &pb.ListVMRequest{
		VmName: vmName,
	}
	resp, err := client.ListVM(ctx, request)
	if err != nil {
		return fmt.Errorf("error starting: %w", err)
	}

	fmt.Println(printVMInfo(resp.VmInfo))
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
					&cli.StringFlag{
						Name:     "entry-point",
						Aliases:  []string{"e"},
						Usage:    "Entry point of the VM",
						Required: false,
					},
				},
				Action: func(ctx *cli.Context) error {
					return startVM(ctx.String("server"), ctx.String("name"), ctx.String("entry-point"))
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
			{
				Name:  "destroy-all",
				Usage: "Destroy all VMs",
				Action: func(ctx *cli.Context) error {
					return destroyAllVMs(ctx.String("server"))
				},
			},
			{
				Name:  "list-all",
				Usage: "List all VMs",
				Action: func(ctx *cli.Context) error {
					return listAllVMs(ctx.String("server"))
				},
			},
			{
				Name:  "list",
				Usage: "List VM info",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Aliases:  []string{"n"},
						Usage:    "Name of the VM to destroy",
						Required: true,
					},
				},
				Action: func(ctx *cli.Context) error {
					return listVM(ctx.String("server"), ctx.String("name"))
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
