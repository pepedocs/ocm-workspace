/*
Copyright © 2023 Jose Cueto

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

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

func ocGetConfig(runAsOcUser string) (*ocConfig, error) {
	commandName := "oc"
	var commandArgs []string

	if len(runAsOcUser) > 0 {
		commandName = "sudo"
		commandArgs = []string{"-Eu", runAsOcUser, "oc", "config", "view", "-o", "json"}
	} else {
		commandArgs = []string{"config", "view", "-o", "json"}
	}

	bytes, err := exec.Command(
		commandName,
		commandArgs...).Output()

	if err != nil {
		return nil, err
	}

	var config ocConfig
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// Gets the current OpenShift cluster that a user is logged in.
func ocGetCurrentOcmCluster(runAsOcUser string) (string, error) {
	config, err := ocGetConfig(runAsOcUser)
	if err != nil {
		return "", err
	}
	parts := strings.Split(config.CurrentContext, "/")
	if len(parts) > 2 {
		return parts[1], nil
	}

	return "", nil

}

// Gets the current OpenShift namespace.
func ocGetCurrentNamespace(runAsOcUser string) (string, error) {
	config, err := ocGetConfig(runAsOcUser)
	if err != nil {
		return "", err
	}

	currentContext := config.CurrentContext

	for _, context := range config.Contexts {
		if context.Name == currentContext {
			return context.Context["namespace"], nil
		}
	}

	err = fmt.Errorf("current context not found: %s", currentContext)
	return "", err
}

// Gets free/unused network ports
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
