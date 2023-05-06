package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"

	gocmd "github.com/go-cmd/cmd"
)

type ocContext struct {
	Name    string            `json:"name"`
	Context map[string]string `json:"context"`
}

type ocConfig struct {
	Contexts       []ocContext `json:"contexts"`
	CurrentContext string      `json:"current-context"`
}

var unknownNamespace = "unknownNamespace"

func runCommandListStreamOutput(commandList [][]string) {
	for _, command := range commandList {
		runCommandStreamOutput(command[0], command[1:]...)
	}
}

// Runs a blocking command (go-cmd) and streams its output.
// https://github.com/go-cmd/cmd/blob/master/examples/blocking-streaming/main.go
func runCommandStreamOutput(cmdName string, args ...string) gocmd.Status {
	cmdOptions := gocmd.Options{
		Buffered:  false,
		Streaming: true,
	}
	command := gocmd.NewCmdOptions(cmdOptions, cmdName, args...)

	doneChan := make(chan struct{})

	go func() {
		defer close(doneChan)
		// Done when both channels have been closed
		// https://dave.cheney.net/2013/04/30/curious-channels
		for command.Stdout != nil || command.Stderr != nil {
			select {
			case line, open := <-command.Stdout:
				if !open {
					command.Stdout = nil
					continue
				}
				fmt.Println(line)
			case line, open := <-command.Stderr:
				if !open {
					command.Stderr = nil
					continue
				}
				fmt.Fprintln(os.Stderr, line)
			}
		}
	}()

	// Run and wait for Cmd to return, discard Status
	<-command.Start()
	// Wait for goroutine to print everything
	<-doneChan
	return command.Status()
}

func ocGetCurrentNamespace(runAsOcUser string) string {
	bytes, err := exec.Command("sudo", "-Eu", runAsOcUser, "oc", "config", "view", "-o", "json").Output()
	if err != nil {
		log.Println(err)
		return unknownNamespace
	}
	var config ocConfig
	json.Unmarshal(bytes, &config)

	currentContext := config.CurrentContext

	for _, context := range config.Contexts {
		if context.Name == currentContext {
			return context.Context["namespace"]
		}
	}

	return unknownNamespace
}

func getFreePorts(numPorts int) ([]int, error) {
	var ports []int

	for idx := 0; idx < numPorts; idx++ {
		addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
		if err != nil {
			return ports, err
		}
		listener, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return ports, err
		}

		defer listener.Close()
		ports = append(ports, listener.Addr().(*net.TCPAddr).Port)
	}
	return ports, nil

}
