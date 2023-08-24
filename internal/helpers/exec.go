package internal

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	gocmd "github.com/go-cmd/cmd"
	logger "github.com/sirupsen/logrus"
)

func RunCommandBackground(cmdName string, cmdArgs []string, envVars [][]string) error {
	cmd := exec.Command(cmdName, cmdArgs...)
	for _, env := range envVars {
		envVar := fmt.Sprintf("%s=%s", env[0], env[1])
		cmd.Env = append(cmd.Env, envVar)
	}

	err := cmd.Start()

	if err != nil {
		return err
	}

	go func() {
		err = cmd.Wait()
		if err != nil {
			logger.Errorf("Command finished with error: %v", err)
		}
	}()

	return nil
}

func RunCommandOutput(cmdName string, cmdArgs ...string) ([]byte, error) {
	// log.Printf("Running command: %s %s\n", cmdName, cmdArgs)
	cmd := exec.Command(cmdName, cmdArgs...)
	return cmd.Output()
}

func RunCommandPipeStdin(cmdName string, cmdArgs ...string) ([]byte, error) {
	// log.Printf("Running command: %s %s\n", cmdName, cmdArgs)
	cmd := exec.Command(cmdName, cmdArgs...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, "")
	}()

	return cmd.Output()
}

func RunCommandWithOsFiles(cmdName string, stdout *os.File, stderr *os.File, stdin *os.File, cmdArgs ...string) error {
	// log.Printf("Running command: %s %s\n", cmdName, cmdArgs)
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdout = stdout
	cmd.Stdin = stdin
	cmd.Stderr = stderr
	err := cmd.Run()
	return err
}

func RunCommandListStreamOutput(commandList [][]string) []error {
	errors := []error{}
	for _, command := range commandList {
		status := RunCommandStreamOutput(command[0], command[1:]...)
		if status.Error != nil {
			logger.Errorf("Failed to run command \"command\": %v", status.Error)
			errors = append(errors, status.Error)
		}
	}
	return errors
}

// Runs a blocking command (go-cmd) and streams its output.
// https://github.com/go-cmd/cmd/blob/master/examples/blocking-streaming/main.go
func RunCommandStreamOutput(cmdName string, args ...string) gocmd.Status {
	// log.Printf("Running command: %s %s\n", cmdName, args)

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
