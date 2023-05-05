package cmd

import (
	"fmt"
	"os"

	gocmd "github.com/go-cmd/cmd"
)

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
