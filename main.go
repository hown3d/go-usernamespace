package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/moby/moby/pkg/reexec"
)

func init() {
	reexec.Register("container-hello", containerHello)
	// inside child process, this will return true
	if reexec.Init() {
		fmt.Println("init func was called from reexec, exiting...")
		os.Exit(0)
	}
}

func containerHello() {
	hello("container")
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
	cmd.Stdout = os.Stdout

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUSER,
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getuid(),
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getgid(),
				Size:        1,
			},
		},
	}

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