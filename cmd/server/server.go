package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/abshkbh/chv-lambda/openapi"
	"github.com/abshkbh/chv-lambda/out/protos"
	"google.golang.org/grpc"
)

const (
	binPath         = "/home/maverick/projects/chv-lambda/resources/bin"
	numBootVcpus    = 1
	memorySizeBytes = 512 * 1024 * 1024
	// Case sensitive.
	serialPortMode = "Tty"
	// Case sensitive.
	consolePortMode = "Off"
	chvBinPath      = "/home/maverick/projects/chv-lambda/resources/bin/cloud-hypervisor"

	tapDeviceName           = "tap0"
	numNetDeviceQueues      = 2
	netDeviceQueueSizeBytes = 256
	netDeviceId             = "_net0"
	stateDir                = "/run/chv-lambda"
)

var (
	kernelPath    = binPath + "/compiled-vmlinux.bin"
	rootfsPath    = binPath + "/ext4.img"
	initPath      = "/bin/bash"
	kernelCmdline = "console=ttyS0 root=/dev/vda rw init=" + initPath
)

func String(s string) *string {
	return &s
}

func Int32(i int32) *int32 {
	return &i
}

// runCloudHypervisor starts the chv binary at `chvBinPath` on the given `apiSocket`.
func runCloudHypervisor(chvBinPath string, apiSocketPath string) error {
	cmd := exec.Command(chvBinPath, "--api-socket", apiSocketPath)
	cmd.Stdout = log.StandardLogger().Writer()
	cmd.Stderr = log.StandardLogger().Writer()

	// Run the command
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error spawning chv binary: %w", err)
	}

	log.Println("Spawn successful")
	return nil
}

func getVmStateDir(vmName string) string {
	return path.Join(stateDir, vmName)
}

func getVmSocketPath(vmStateDir string, vmName string) string {
	return path.Join(vmStateDir, vmName+".sock")
}

func unixSocketClient(socketPath string) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
		Timeout: time.Second * 30,
	}
}

func createApiClient(apiSocketPath string) *openapi.APIClient {
	configuration := openapi.NewConfiguration()
	configuration.HTTPClient = unixSocketClient(apiSocketPath)
	configuration.Servers = openapi.ServerConfigurations{
		{
			URL: "http://localhost/api/v1",
		},
	}
	return openapi.NewAPIClient(configuration)
}

func waitForServer(ctx context.Context, apiClient *openapi.APIClient, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			default:
				resp, r, err := apiClient.DefaultAPI.VmmPingGet(ctx).Execute()
				if err == nil {
					log.WithFields(log.Fields{
						"buildVersion": *resp.BuildVersion,
						"statusCode":   r.StatusCode,
					}).Info("cloud-hypervisor server up")
					errCh <- nil
					return
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	return <-errCh
}

func createVM(ctx context.Context, vmName string) error {
	vmStateDir := getVmStateDir(vmName)
	err := os.MkdirAll(vmStateDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create vm state dir for vm: %s", vmName)
	}

	apiSocketPath := getVmSocketPath(vmStateDir, vmName)
	go func() {
		err := runCloudHypervisor(chvBinPath, apiSocketPath)
		if err != nil {
			log.WithError(err).Fatal("failed to spawn cloud-hypervisor server")
		}
	}()

	apiClient := createApiClient(apiSocketPath)
	err = waitForServer(context.Background(), apiClient, 10*time.Second)
	if err != nil {
		log.WithError(err).Fatal("failed to wait for cloud-hypervisor server")
	}

	// Create a new VM configuration
	vmConfig := openapi.VmConfig{
		Payload: openapi.PayloadConfig{
			Kernel:  &kernelPath,
			Cmdline: &kernelCmdline,
		},
		Disks:   []openapi.DiskConfig{{Path: rootfsPath}},
		Cpus:    &openapi.CpusConfig{BootVcpus: numBootVcpus, MaxVcpus: numBootVcpus},
		Memory:  &openapi.MemoryConfig{Size: memorySizeBytes},
		Serial:  openapi.NewConsoleConfig(serialPortMode),
		Console: openapi.NewConsoleConfig(consolePortMode),
		Net:     []openapi.NetConfig{{Tap: String(tapDeviceName), NumQueues: Int32(numNetDeviceQueues), QueueSize: Int32(netDeviceQueueSizeBytes), Id: String(netDeviceId)}},
	}

	req := apiClient.DefaultAPI.CreateVM(ctx)
	req = req.VmConfig(vmConfig)

	// Execute the request
	resp, err := req.Execute()
	if err != nil {
		return fmt.Errorf("failed to start VM: %w %v", err, resp.Body)
	}

	log.Infof("create resp: %v", resp)
	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return fmt.Errorf("failed to start VM. bad status: %v", resp)
	}

	log.Infof("before bootVM")
	resp, err = apiClient.DefaultAPI.BootVM(ctx).Execute()
	if err != nil {
		return fmt.Errorf("failed to boot VM: %w %v", err, resp.Body)
	}

	log.Infof("boot resp: %v", resp)
	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return fmt.Errorf("failed to boot VM. bad status: %v", resp)
	}

	return nil
}

type server struct {
	protos.UnimplementedVMManagementServiceServer
}

func (s *server) StartVM(ctx context.Context, req *protos.VMRequest) (*protos.VMResponse, error) {
	log.Infof("received request to start VM: %v", req.GetVmName())
	vmName := req.GetVmName()
	err := createVM(ctx, vmName)

	if err != nil {
		return nil, err
	}

	return &protos.VMResponse{}, nil
}

func (s *server) StopVM(ctx context.Context, req *protos.VMRequest) (*protos.VMResponse, error) {
	log.Printf("Received request to stop VM: %v", req.GetVmName())
	// Implement your VM stop logic here
	return &protos.VMResponse{}, nil
}

func (s *server) DestroyVM(ctx context.Context, req *protos.VMRequest) (*protos.VMResponse, error) {
	log.Printf("Received request to destroy VM: %v", req.GetVmName())
	// Implement your VM destroy logic here
	return &protos.VMResponse{}, nil
}

func main() {
	err := os.MkdirAll(stateDir, 0755)
	if err != nil {
		log.WithError(err).Fatal("failed to create vm state dir")
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.WithError(err).Fatalf("failed to listen")
	}
	s := grpc.NewServer()

	protos.RegisterVMManagementServiceServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.WithError(err).Fatalf("failed to serve")
	}
}