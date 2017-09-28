package container

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/gustavosbarreto/go-microcontainer/rootfs/alpine"
	"github.com/gustavosbarreto/go-microcontainer/rootfs/scratch"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var rootfs RootFSProvider

var procAttr = &syscall.SysProcAttr{
	Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWIPC,
}

func CreateNetworkNamespace() {
	procAttr.Cloneflags |= syscall.CLONE_NEWNET
}

func SetRootFSProvider(provider string) {
	switch provider {
	case "alpine":
		rootfs = alpine.NewRootFS()
		break
	case "scratch":
		rootfs = scratch.NewRootFS()
	}
}

func Main(main func()) {
	if os.Getenv("CONTAINER_RUN") != "true" {
		if err := rootfs.PullRootFS(); err != nil {
			logrus.Error(errors.Wrapf(err, "failed to pull rootfs"))
			os.Exit(1)
		}

		cmd := exec.Command("/proc/self/exe", os.Args...)
		cmd.Env = []string{"CONTAINER_RUN=true"}

		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout

		cmd.SysProcAttr = procAttr

		if err := cmd.Run(); err != nil {
			logrus.Error(errors.Wrapf(err, "failed to exec child process"))
		}

		cmd.Wait()
	} else {
		if err := syscall.Mount("", "/", "", syscall.MS_SLAVE|syscall.MS_REC, ""); err != nil {
			logrus.Error(errors.Wrapf(err, "mount propagate failed"))
			os.Exit(1)
		}

		if err := syscall.Mount("rootfs", "rootfs", "", syscall.MS_BIND, ""); err != nil {
			logrus.Error(errors.Wrapf(err, "rootfs bind failed"))
			os.Exit(1)
		}

		if err := syscall.PivotRoot("rootfs", "rootfs/tmp"); err != nil {
			logrus.Error(errors.Wrapf(err, "pivot_root failed"))
			os.Exit(1)
		}

		if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
			logrus.Error(errors.Wrapf(err, "mount proc failed"))
			os.Exit(1)
		}

		main()
	}
}
