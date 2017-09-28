package main

import (
	"os"
	"os/exec"

	container "github.com/gustavosbarreto/go-microcontainer"
)

func main() {
	container.SetRootFSProvider("alpine")
	container.Main(func() {
		cmd := exec.Command("/bin/sh")

		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout

		cmd.Start()
		cmd.Wait()
	})
}
