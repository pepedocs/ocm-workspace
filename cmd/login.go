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

	"github.com/golang/glog"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Represents a port bind between a k8s service, a container (forwarded port) and a container's host
type serviceHostPortBind struct {
	HostPort      int    `yaml:"hostPort"`
	ServicePort   int    `yaml:"servicePort"`
	ServiceName   string `yaml:"serviceName"`
	ParentService string `yaml:"parentService"`
	SourceKind    string `yaml:"sourceKind"`
	SourceName    string `yaml:"sourceName"`
	Namespace     string `yaml:"namespace"`
}

type serviceHostPortBindList struct {
	ServiceHostPortBinds []serviceHostPortBind `yaml:"serviceHostPortBinds"`
}

// Represents a k8s service's port forward source (e.g. pod or deployment)
type serviceRefForwardPortSource struct {
	Kind string `mapstructure:"kind"`
	Name string `mapstructure:"name"`
}

// Represents a k8s service port forward parameters
type serviceRefForwardPort struct {
	Name      string                      `mapstructure:"name"`
	Namespace string                      `mapstructure:"namespace"`
	Source    serviceRefForwardPortSource `mapstructure:"source"`
	Port      int                         `mapstructure:"port"`
}

// Represents a parent service's list of child services that need port forwards
type serviceRefConfig struct {
	Name         string                  `mapstructure:"name"`
	ForwardPorts []serviceRefForwardPort `mapstructure:"forwardPorts"`
}

type ocmWorkspaceConfig struct {
	Services []serviceRefConfig `mapstructure:"services"`
}

var (
	loginCmdArgs struct {
		cluster        string
		ocmEnvironment string
		service        string
		isOcmLoginOnly bool
	}
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Runs the ocm workspace container and logs in to a cluster.",
	Long: `Runs the ocm workspace container with the ocm workspace as its entrypoint.
           The entrypoint is supplied with the "clusterLogin" flag to proceed with
		   logging into a cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		ocmEnvironment := "production"
		serviceRef := loginCmdArgs.service
		if len(loginCmdArgs.ocmEnvironment) > 0 {
			ocmEnvironment = loginCmdArgs.ocmEnvironment
		}
		runOCMWorkspaceContainer(
			loginCmdArgs.cluster,
			ocmEnvironment,
			serviceRef,
			loginCmdArgs.isOcmLoginOnly)
	},
}

// Runs the OCM Workspace container using the image built by the "build" command
func runOCMWorkspaceContainer(
	ocmCluster string,
	ocmEnvironment string,
	serviceRef string,
	isOcmLoginOnly bool) {
	envVarOcmUser := fmt.Sprintf("OCM_USER=%s", viper.GetString("ocUser"))
	envVarOcmToken := fmt.Sprintf("OCM_TOKEN=%s", viper.GetString("ocmToken"))
	envVarCluster := fmt.Sprintf("OCM_CLUSTER=%s", ocmCluster)
	envVarIsOCMLoginOnly := fmt.Sprintf("IS_OCM_LOGIN_ONLY=%v", isOcmLoginOnly)
	userHome := fmt.Sprintf("/home/%s", viper.GetString("ocUser"))

	// Paths to where these files are mounted in the workspace container
	containerBackplaneConfigPath := "/backplane-config.json"
	ocmWorkspaceConfigPath := "/.ocm-workspace.yaml"

	volMapBackplaneConfig := fmt.Sprintf("%s/.config/backplane/config.prod.json:%s:ro", userHome, containerBackplaneConfigPath)
	if ocmEnvironment == "staging" {
		volMapBackplaneConfig = fmt.Sprintf("%s/.config/backplane/config.stage.json:%s:ro", userHome, containerBackplaneConfigPath)
	}

	volMapTerminalDir := "./terminal:/terminal:ro"
	volMapOcmWorkspaceConfig := fmt.Sprintf("%s/.ocm-workspace.yaml:%s:ro", userHome, ocmWorkspaceConfigPath)
	envVarOcmEnvironment := fmt.Sprintf("OCM_ENVIRONMENT=%s", ocmEnvironment)
	envVarBackplaneConfig := fmt.Sprintf("BACKPLANE_CONFIG=%s", containerBackplaneConfigPath)
	suffix := uuid.New()
	containerName := fmt.Sprintf("ow-%s-%s", ocmCluster, suffix.String()[:6])
	commandArgs := []string{
		"run",
		"--name",
		containerName,
		"-it",
		"--privileged",
		"-e",
		envVarOcmUser,
		"-e",
		envVarOcmEnvironment,
		"-e",
		envVarOcmToken,
		"-e",
		envVarCluster,
		"-e",
		envVarBackplaneConfig,
		"-e",
		envVarIsOCMLoginOnly,
		"-v",
		volMapBackplaneConfig,
		"-v",
		volMapTerminalDir,
		"-v",
		volMapOcmWorkspaceConfig,
	}

	var config ocmWorkspaceConfig
	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatal("Failed to unmarshal config: ", err)
	}

	// Allocate free port and map host port for OpenShift console
	ports, err := getFreePorts(1)
	if err != nil {
		glog.Fatal("Failed to generate port for Openshift console: ", err)
	}

	portStr := strconv.Itoa(ports[0])
	envVarOpenShiftConsolePort := fmt.Sprintf("OPENSHIFT_CONSOLE_PORT=%s", portStr)
	commandArgs = append(commandArgs, "-e")
	commandArgs = append(commandArgs, envVarOpenShiftConsolePort)
	openShiftConsolePortMap := fmt.Sprintf("127.0.0.1:%s:%s", portStr, portStr)
	commandArgs = append(commandArgs, "-p")
	commandArgs = append(commandArgs, openShiftConsolePortMap)

	var portBindList serviceHostPortBindList

	// Bind service ports to free host ports
	for _, svcRefConf := range config.Services {
		if svcRefConf.Name != serviceRef {
			continue
		}
		for _, forwardPort := range svcRefConf.ForwardPorts {
			ports, err := getFreePorts(1)
			if err != nil {
				log.Fatalf("Failed to generate port for %s: %s", svcRefConf.Name, err)
			}
			// Host port must only be locally accessible
			portMap := fmt.Sprintf("127.0.0.1:%s:%s", strconv.Itoa(ports[0]), strconv.Itoa(ports[0]))
			commandArgs = append(commandArgs, "-p")
			commandArgs = append(commandArgs, portMap)

			var svcHostPort serviceHostPortBind
			svcHostPort.HostPort = ports[0]
			svcHostPort.ServiceName = forwardPort.Name
			svcHostPort.ServicePort = forwardPort.Port
			svcHostPort.ParentService = svcRefConf.Name
			svcHostPort.SourceKind = forwardPort.Source.Kind
			svcHostPort.SourceName = forwardPort.Source.Name
			svcHostPort.Namespace = forwardPort.Namespace
			portBindList.ServiceHostPortBinds = append(
				portBindList.ServiceHostPortBinds,
				svcHostPort,
			)
		}
	}

	// Store the service and host port binds into an environment variable
	portBindListYamlStr, err := yaml.Marshal(portBindList)
	if err != nil {
		log.Printf("Failed to marshal port bind list string %s: %s", portBindListYamlStr, err)
	}
	envVarPortBindListYamlStr := fmt.Sprintf("HOST_PORT_MAPS=\"%s\"", portBindListYamlStr)
	commandArgs = append(
		commandArgs,
		"-e",
		envVarPortBindListYamlStr,
		"--entrypoint",
		"./workspace",
		"ocm-workspace:latest",
		"clusterLogin",
		ocmCluster,
		"--config",
		ocmWorkspaceConfigPath,
	)
	err = runCommandWithOsFiles("podman", os.Stdout, os.Stderr, os.Stdin, commandArgs...)
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
