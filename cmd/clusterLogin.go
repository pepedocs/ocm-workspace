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
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	pkgInt "ocm-workspace/internal"
	pkgIntHelper "ocm-workspace/internal/helpers"

	logger "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ocmWorkspace *ocmWorkspaceContainer

var clusterLoginCmd = &cobra.Command{
	Use:    "clusterLogin",
	Short:  "Logs in to an OpenShift Dedicated cluster.",
	PreRun: toggleDebug,
	Run:    onClusterLogin,
}

func onClusterLogin(cmd *cobra.Command, args []string) {

	// Use ocm workspace container instance for in-container commands
	ocmWorkspace = NewOcmWorkspaceContainer(config)
	if err := checkContainerCommand(); err != nil {
		logger.Fatal(err)
	}

	configureOCMUser()
	configureWorkspaceDirs()
	OCMLogin()
	OCMBackplaneLogin()

	customPortMapsStr := strings.Trim(getEnvVar("CUSTOM_PORT_MAPS"), ",")
	var allocatedContainerPorts []string

	for _, pm := range strings.Split(customPortMapsStr, ",") {
		ports := strings.Split(pm, ":")
		allocatedContainerPorts = append(allocatedContainerPorts, ports[1])
	}

	plugins := config.Plugins
	for _, plug := range plugins {

		if plug.RunOn == "ocmBackplaneLoginSuccess" {
			err := OCMBackplaneLoginSuccess(plug, allocatedContainerPorts)
			if err != nil {
				logger.Fatalf("Plugin %v failed to run: %v", plug.Name, err)
			}
		}

		if (len(allocatedContainerPorts) - plug.AllocatePorts) > 1 {
			allocatedContainerPorts = allocatedContainerPorts[plug.AllocatePorts+1:]
		}
	}

	runTerminal()
}

func runTerminal() {
	// Run terminal
	status := pkgIntHelper.RunCommandStreamOutput("cp", "/terminal/bashrc", ocmWorkspace.UserBashrcPath)
	if status.Exit != 0 {
		logger.Fatalf("Failed to copy /terminal/bashrc: %v", status.Error)
	}

	file, err := os.OpenFile(ocmWorkspace.UserBashrcPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		logger.Errorf("Failed to open %s: %s\n", ocmWorkspace.UserBashrcPath, err)
	} else {
		defer file.Close()

		ps1String := fmt.Sprintf(
			"\nPS1='[%s %s $(/usr/bin/workspace currentCluster) $(/usr/bin/workspace currentNamespace -u %s)]$ '\n",
			ocmWorkspace.HostUser,
			ocmWorkspace.OcmEnvironment,
			ocmWorkspace.HostUser)
		_, err = file.WriteString(ps1String)
		if err != nil {
			logger.Errorf("Failed to write to file %s: %s\n", ocmWorkspace.UserBashrcPath, err)
		}

		exportStr := "\nexport PATH=$PATH"
		for _, path := range config.AddToPATHEnv {
			exportStr += fmt.Sprintf(":%s", path)
		}
		_, err = file.WriteString(exportStr)
		if err != nil {
			logger.Errorf("Failed to write to file %s: %s\n", ocmWorkspace.UserBashrcPath, err)
		}

		for _, path := range config.ExportEnvVars {
			exportStr = fmt.Sprintf("\nexport %s", path)
			_, err = file.WriteString(exportStr)
			if err != nil {
				logger.Errorf("Failed to write to file %s: %s\n", ocmWorkspace.UserBashrcPath, err)
			}
		}
	}
	err = pkgIntHelper.RunCommandWithOsFiles("sudo", os.Stdout, os.Stderr, os.Stdin, "-Eu", ocmWorkspace.HostUser, "bash")
	if err != nil {
		logger.Fatal("Failed to run command: ", err)
	}
}

func runPlugin(plug pkgInt.Plugin, configPath string, envVars [][]string) error {
	executable := filepath.Base(plug.ExecPath)
	cmdArgs := []string{
		"-Eu",
		getEnvVar("HOST_USER"),
		fmt.Sprintf("/usr/bin/%s", executable),
		plug.ExecCommand,
		"--config", configPath}
	if debug {
		cmdArgs = append(cmdArgs, "-d")
	}
	logger.Debugf("Running plugin with args: %v %v", cmdArgs, envVars)

	err := pkgIntHelper.RunCommandBackground("sudo", cmdArgs, envVars)
	return err
}

func OCMBackplaneLoginSuccess(plug pkgInt.Plugin, allocatedContainerPorts []string) error {

	// Create (overwrite) plugin config
	configPath := fmt.Sprintf("%s/.%s.yaml", ocmWorkspace.UserHome, plug.Name)
	createConfigFile(configPath, plug.Config)

	if plug.AllocatePorts > len(allocatedContainerPorts) {
		logger.Fatalf("Not enough custom port maps for plugin use.")
	}

	// Pass allocated container ports to plugin as environment variables
	envVars := [][]string{
		{"PLUGIN_SERVICE", getEnvVar("PLUGIN_SERVICE")},
		{"HOST_USER", getEnvVar("HOST_USER")},
	}

	return runPlugin(plug, configPath, envVars)
}

func OCMBackplaneLogin() {
	isOcmLoginOnly, err := strconv.ParseBool(ocmWorkspace.IsOcmLoginOnly)
	if err != nil {
		logger.Fatal("Failed to parse environment variable: ", err)
	}

	if !isOcmLoginOnly {
		// Backplane login
		status := pkgIntHelper.RunCommandStreamOutput(
			"sudo",
			"-Eu",
			ocmWorkspace.HostUser,
			"ocm",
			"backplane",
			"login",
			ocmWorkspace.OcmCluster,
		)

		if status.Exit != 0 {
			logger.Fatalf("OCM backplane login failed: %v", status.Error)
		}
		logger.Info("OCM backplane login successful.")
	}

}

func OCMLogin() {
	logger.Info("Logging into ocm ", ocmWorkspace.OcmEnvironment)

	status := pkgIntHelper.RunCommandStreamOutput(
		"sudo",
		"-Eu",
		ocmWorkspace.HostUser,
		"ocm",
		"login",
		fmt.Sprintf("--token=%s", ocmWorkspace.OcmToken),
		fmt.Sprintf("--url=%s", ocmWorkspace.OcmEnvironment),
	)

	if status.Exit != 0 {
		logger.Fatalf("OCM Login failed: %v", status.Error)
	}

	logger.Info("OCM Login successful.")
}

func configureWorkspaceDirs() {
	// Configure directories
	commands := [][]string{
		{
			"mkdir",
			"-p",
			fmt.Sprintf("%s/.kube", ocmWorkspace.UserHome),
		},
		{
			"chown",
			"-R",
			fmt.Sprintf("%s:%s", ocmWorkspace.HostUser, ocmWorkspace.HostUser),
			fmt.Sprintf("%s/.kube", ocmWorkspace.UserHome),
		},
		{
			"mkdir",
			"-p",
			fmt.Sprintf("%s/.config/ocm", ocmWorkspace.UserHome),
		},
		{
			"chown",
			"-R",
			fmt.Sprintf("%s:%s", ocmWorkspace.HostUser, ocmWorkspace.HostUser),
			fmt.Sprintf("%s/.config/ocm", ocmWorkspace.UserHome),
		},
		{
			"chmod",
			"o+rwx",
			"/ocm-workspace",
		},
	}
	errors := pkgIntHelper.RunCommandListStreamOutput(commands)

	if len(errors) > 0 {
		logger.Fatalf("Encountered errors while configuring workspace directories: %v", errors)
	}

}

func configureOCMUser() {

	// Configure ocm user
	commands := [][]string{
		{
			"useradd",
			"-m",
			ocmWorkspace.HostUser,
			"-d",
			ocmWorkspace.UserHome,
		},
		{
			"usermod",
			"-aG",
			"wheel",
			ocmWorkspace.HostUser,
		},
	}
	errors := pkgIntHelper.RunCommandListStreamOutput(commands)

	if len(errors) > 0 {
		logger.Fatalf("Encountered errors while configuring OCM user: %v", errors)
	}

	line := []byte("\n%wheel         ALL = (ALL) NOPASSWD: ALL\n")
	os.WriteFile("/etc/sudoer", line, 0644)

}

func init() {
	rootCmd.AddCommand(clusterLoginCmd)
}
