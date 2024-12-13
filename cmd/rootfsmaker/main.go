package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/urfave/cli/v2"
	"gvisor.dev/gvisor/pkg/cleanup"
)

const (
	outputDir           = "out"
	workingDir          = outputDir + "/rootfsmaker-working-dir"
	dockerFile          = "./resources/scripts/rootfs/Dockerfile"
	dockerImageName     = "chv-lambda-img"
	dockerContainerName = "chv-lambda"
	rootfsTar           = outputDir + "/rootfs.tar"
	rootfsDir           = outputDir + "/rootfs"
	rootfsExt4Image     = outputDir + "/rootfs-ext4.img"
	mountDir            = outputDir + "/mnt"
	diskSizeInMB        = 2048
)

// runCmd runs `cmdName` with `args`.
func runCmd(cmdName string, args ...string) error {
	cmd := exec.Command(cmdName, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func createRootfsFromDockerfile(dockerFile string) (retErr error) {
	cleanup := cleanup.Make(func() {
		if retErr == nil {
			log.Info("create rootfs from docker file finished")
		}
	})

	defer func() {
		cleanup.Clean()
	}()

	log.Info("creating working dir")
	err := runCmd("mkdir", "-p", workingDir)
	if err != nil {
		return fmt.Errorf("failed to create working dir: %s: %w", workingDir, err)
	}
	cleanup.Add(func() {
		if err := runCmd("rm", "-rf", workingDir); err != nil {
			log.WithError(err).Errorf("failed to cleanup working dir: %s", workingDir)
		}
	})

	log.Info("copying dockerfile")
	srcDockerfile := dockerFile
	dstDockerfile := workingDir + "/Dockerfile"
	err = runCmd("cp", srcDockerfile, dstDockerfile)
	if err != nil {
		return fmt.Errorf(
			"failed to copy docker file: %s to: %s: %w",
			srcDockerfile,
			dstDockerfile,
			err,
		)
	}

	log.Info("creating output folder")
	err = runCmd("mkdir", "-p", outputDir)
	if err != nil {
		return fmt.Errorf("failed to create output folder: %w", err)
	}

	log.Info("building docker image")
	err = runCmd("docker", "build", "-f", dockerFile, "-t", dockerImageName, ".")
	if err != nil {
		return fmt.Errorf("failed to build docker container image: %w", err)
	}

	log.Info("creating container")
	err = runCmd("docker", "create", "--name", dockerContainerName, dockerImageName)
	if err != nil {
		return fmt.Errorf("failed to build docker container: %w", err)
	}

	log.Info("exporting rootfs to tar file")
	err = runCmd("docker", "export", "--output", rootfsTar, dockerContainerName)
	if err != nil {
		return fmt.Errorf("failed to export docker container: %w", err)
	}
	cleanup.Add(func() {
		if err := runCmd("docker", "rm", dockerContainerName); err != nil {
			log.WithError(err).Errorf(
				"failed to cleanup docker container: %s",
				dockerContainerName,
			)
		}
	})

	log.Info("creating img file")
	err = runCmd(
		"dd",
		"if=/dev/zero",
		fmt.Sprintf(
			"of=%s", rootfsExt4Image),
		"bs=1M",
		fmt.Sprintf("count=%s", strconv.Itoa(diskSizeInMB)),
	)
	if err != nil {
		return fmt.Errorf("failed to create img file: %w", err)
	}

	log.Info("formatting img file to ext4")
	err = runCmd("mkfs.ext4", rootfsExt4Image)
	if err != nil {
		return fmt.Errorf("failed to format ext4 image: %w", err)
	}

	log.Info("creating loopback mount directory")
	err = runCmd("mkdir", "-p", mountDir)
	if err != nil {
		return fmt.Errorf("failed to create loopback mount dir: %w", err)
	}
	cleanup.Add(func() {
		if err := runCmd("rm", "-rf", mountDir); err != nil {
			log.WithError(err).Errorf("failed to cleanup mount dir: %s", mountDir)
		}
	})

	log.Info("mounting loopback mount")
	err = runCmd("mount", "-o", "loop", rootfsExt4Image, mountDir)
	if err != nil {
		return fmt.Errorf("failed to mount ext4 image: %w", err)
	}
	cleanup.Add(func() {
		if err := runCmd("umount", mountDir); err != nil {
			log.WithError(err).Errorf("failed to umount dir: %s", mountDir)
		}
	})

	log.Info("extracting rootfs tar to mount dir")
	err = runCmd("tar", "-xvf", rootfsTar, "-C", mountDir)
	if err != nil {
		return fmt.Errorf("failed to extract rootfs tar to mount dir: %w", err)
	}
	cleanup.Add(func() {
		if err := runCmd("rm", "-rf", rootfsTar); err != nil {
			log.WithError(err).Errorf("failed to delete rootfs tar: %s", rootfsTar)
		}
	})

	return nil
}

func main() {
	log.Info("making guest rootfs")

	app := &cli.App{
		Name:  "rootfsmaker",
		Usage: "A script to make a guest rootfs for a VM",
		Commands: []*cli.Command{
			{
				Name:  "create",
				Usage: "Create guest rootfs from docker file",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "dockerfile",
						Aliases:  []string{"d"},
						Usage:    "Path to the docker file",
						Required: true,
					},
				},
				Action: func(ctx *cli.Context) error {
					return createRootfsFromDockerfile(ctx.String("dockerfile"))
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.WithError(err).Fatal("failed to create rootfs")
	}
}