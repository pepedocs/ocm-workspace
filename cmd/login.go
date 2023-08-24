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
	"log"
	"os"
	"strconv"

	"github.com/google/uuid"
	logger "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	pkgInt "ocm-workspace/internal"
	pkgIntHelper "ocm-workspace/internal/helpers"
)

var (
	loginCmdArgs struct {
		cluster        string
		ocmEnvironment string
		service        string
		isOcmLoginOnly bool
	}
)

var loginCmd = &cobra.Command{
	Use:    "login",
	Short:  "Launches the ocm-workspace container and logs into a cluster.",
	PreRun: toggleDebug,
	Run:    onLogin,
}

// Runs the OCM Workspace container using the image built by the "build" command
func onLogin(cmd *cobra.Command, args []string) {
	ocmEnvironment := "production"

	if len(loginCmdArgs.ocmEnvironment) > 0 {
		ocmEnvironment = loginCmdArgs.ocmEnvironment
	}

	ceFactory := pkgInt.NewCeFactory(map[string]interface{}{
		"ceName": "podman",
	})

	ocmCluster := loginCmdArgs.cluster
	isOcmLoginOnly := loginCmdArgs.isOcmLoginOnly

	ce, err := ceFactory.Create()
	if err != nil {
		logger.Fatal("Failed to create container engine: ", err)
	}

	ocmLongLivedTokenPath := config.GetOcmLongLivedTokenPath()
	var ocmToken string

	if len(ocmLongLivedTokenPath) > 0 {
		content, err := os.ReadFile(ocmLongLivedTokenPath)
		if err != nil {
			logger.Fatalf("Failed to open long lived token file: %v", content)
		}
		ocmToken = string(content)
	} else {
		ocmToken, err = pkgIntHelper.OcmGetOCMToken()
		if err != nil {
			logger.Fatal("Failed to fetch the OCM token: ", err)
		}
	}

	// Path where backplane config is mounted in the container
	containerBackplaneConfigPath := "/backplane-config.json"
	// Path where workspace config is mounted in the container
	ocmWorkspaceConfigPath := "/.ocm-workspace.yaml"

	// Allocate free port and map host port for OpenShift console
	ports, err := pkgIntHelper.GetFreePorts(1)
	if err != nil {
		logger.Fatal("Failed to generate port for Openshift console: ", err)
	}
	openshiftConsolePort := strconv.Itoa(ports[0])

	// Gather values for the container's environment variables
	ce.AppendEnvVar("HOST_USER", config.GetHostUser())
	ce.AppendEnvVar("OC_USER", config.GetOcUser())
	ce.AppendEnvVar("OCM_CLUSTER", ocmCluster)
	ce.AppendEnvVar("IS_OCM_LOGIN_ONLY", strconv.FormatBool(isOcmLoginOnly))
	ce.AppendEnvVar("OCM_TOKEN", ocmToken)
	ce.AppendEnvVar("IS_IN_CONTAINER", "true")
	ce.AppendEnvVar("OCM_ENVIRONMENT", ocmEnvironment)
	ce.AppendEnvVar("BACKPLANE_CONFIG", containerBackplaneConfigPath)
	ce.AppendEnvVar("OPENSHIFT_CONSOLE_PORT", openshiftConsolePort)
	ce.AppendEnvVar("PLUGIN_SERVICE", loginCmdArgs.service)

	// Gather values for the container's host-mounted volumes
	if ocmEnvironment == "production" {
		ce.AppendVolMap(
			fmt.Sprintf("%s/.config/backplane/%s", config.GetUserHome(), config.GetBackplaneConfigProd()),
			containerBackplaneConfigPath,
			"ro",
		)
	} else {
		ce.AppendVolMap(
			fmt.Sprintf("%s/.config/backplane/%s", config.GetUserHome(), config.GetBackplaneConfigStage()),
			containerBackplaneConfigPath,
			"ro",
		)
	}
	ce.AppendVolMap("./terminal", "/terminal", "ro")
	ce.AppendVolMap(fmt.Sprintf("%s/.ocm-workspace.yaml", config.GetUserHome()), ocmWorkspaceConfigPath, "ro")

	for _, dirMap := range config.GetCustomDirMaps() {
		ce.AppendVolMap(dirMap.HostDir, dirMap.ContainerDir, dirMap.FileAttrs)
	}

	// Mount plugin executables
	plugins := config.GetPlugins()
	for _, plug := range plugins {
		ce.AppendVolMap(plug.ExecPath, fmt.Sprintf("/%s", plug.Name), "ro")
	}

	// Gather values for the containers host-mapped TCP ports
	// Openshift console port
	ce.AppendPortMap(openshiftConsolePort, openshiftConsolePort, "127.0.0.1")

	var customPortMaps string
	for _, pm := range config.CustomPortMaps {
		ports, err = pkgIntHelper.GetFreePorts(1)
		if err != nil {
			logger.Fatalf("Failed to allocate host ports: %v", err)
		}
		pm.HostPort = strconv.Itoa(ports[0])
		customPortMaps += fmt.Sprintf("%s:%s,", pm.HostPort, pm.ContainerPort)
		ce.AppendPortMap(pm.HostPort, pm.ContainerPort, "127.0.0.1")
	}
	ce.AppendEnvVar("CUSTOM_PORT_MAPS", customPortMaps)

	suffix := uuid.New()
	containerName := fmt.Sprintf("ow-%s-%s", ocmCluster, suffix.String()[:6])

	runCmd := ce.GetRunArgs(
		containerName,
		"./workspace",
		"ocm-workspace:latest",
		"clusterLogin",
		ocmCluster,
		"--config",
		ocmWorkspaceConfigPath,
	)

	if debug {
		runCmd = append(runCmd, "-d")
	}

	logger.Debugf("Container run command: %v", runCmd)

	err = pkgIntHelper.RunCommandWithOsFiles(
		ce.GetExecName(),
		os.Stdout,
		os.Stderr,
		os.Stdin,
		runCmd...,
	)

	if err != nil {
		log.Fatal("Failed to run command: ", err)
	}
}

func init() {
	rootCmd.AddCommand(loginCmd)

	flags := loginCmd.Flags()
	flags.StringVarP(
		&loginCmdArgs.cluster,
		"ocmCluster",
		"c",
		"",
		"Cluster name or id.",
	)

	flags.StringVarP(
		&loginCmdArgs.ocmEnvironment,
		"ocmEnvironment",
		"e",
		"production",
		"OCM environemnt (production, staging)",
	)

	flags.StringVarP(
		&loginCmdArgs.service,
		"service",
		"s",
		"",
		"OpenShift service reference name.",
	)

	flags.BoolVar(
		&loginCmdArgs.isOcmLoginOnly,
		"isOcmLoginOnly",
		false,
		"Log in to OCM only.",
	)
}
