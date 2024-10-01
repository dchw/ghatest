package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	fmt.Println("*** Passthrough ***")

	cmd := exec.Command("./Runner.Worker.Legit", os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("passthrough failed: %s\n", err.Error()) // handle error
	}

	os.Exit(cmd.ProcessState.ExitCode())
}
