package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/docker/docker/pkg/reexec"
	"github.com/syndtr/gocapability/capability"
)

func init() {
	// reexec.Register("container-hello", containerHello(dir))
	reexec.Register("container-hello", containerHello)
	// inside child process, this will return true
	if reexec.Init() {
		fmt.Println("init func was called from reexec, exiting...")
		os.Exit(0)
	}
}

var dir, _ = os.Getwd()

func nsRun() {
	cmd := exec.Command("/bin/sh")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = []string{"PS1=-[ns-process]- # "}

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running the /bin/sh command - %s\n", err)
		os.Exit(1)
	}
}

func containerHello() {
	hello("container")
	nsRun()
}

func hostHello() {
	hello("host")
}

func hello(name string) {
	fmt.Printf("hello from %v\n", name)
	uid := os.Getuid()
	gid := os.Getgid()
	fmt.Printf("I'm running with uid: %v and gid: %v\n", uid, gid)
}

func main() {
	cmd := reexec.Command("container-hello")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// user, err := user.Current()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// mapping, err := idtools.NewIdentityMapping(user.Username)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// uidMappings := make([]syscall.SysProcIDMap, len(mapping.UIDs()))
	// for i, ident := range mapping.UIDs() {
	// 	uidMappings[i] = syscall.SysProcIDMap{
	// 		ContainerID: ident.ContainerID,
	// 		HostID:      ident.HostID,
	// 		Size:        ident.Size,
	// 	}
	// }

	// gidMappings := make([]syscall.SysProcIDMap, len(mapping.GIDs()))
	// for i, ident := range mapping.GIDs() {
	// 	gidMappings[i] = syscall.SysProcIDMap{
	// 		ContainerID: ident.ContainerID,
	// 		HostID:      ident.HostID,
	// 		Size:        ident.Size,
	// 	}
	// }

	sysProcAttr := cmd.SysProcAttr
	// defined in https://github.com/torvalds/linux/blob/master/include/uapi/linux/capability.h
	// add CAP_SETGID and CAP_SETUID
	cmd.SysProcAttr.AmbientCaps = []uintptr{uintptr(capability.CAP_SETUID), uintptr(capability.CAP_SETGID)}
	sysProcAttr.Cloneflags = syscall.CLONE_NEWUSER | syscall.CLONE_NEWNS

	sysProcAttr.UidMappings = []syscall.SysProcIDMap{
		{
			ContainerID: 0,
			HostID:      os.Getuid(),
			Size:        1,
		},
	}
	sysProcAttr.GidMappings = []syscall.SysProcIDMap{
		{
			ContainerID: 0,
			HostID:      os.Getgid(),
			Size:        1,
		},
	}

	// https://github.com/golang/go/issues/50098#issuecomment-995478035
	// need to run newuidmap and newgidmap to map more than just the current user
	// sysProcAttr.UidMappings = uidMappings
	// sysProcAttr.GidMappings = gidMappings

	fmt.Println("starting reexec.Command")
	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting the reexec.Command - %s\n", err)
		os.Exit(1)
	}

	hostHello()

	if err := cmd.Wait(); err != nil {
		fmt.Printf("Error waiting for the reexec.Command - %s\n", err)
		os.Exit(1)
	}
}
