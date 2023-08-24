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

	logger "github.com/sirupsen/logrus"

	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	pkgIntHelper "ocm-workspace/internal/helpers"
)

type ocDeploymentContainer struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

type ocDeploymentTemplateSpec struct {
	Containers []ocDeploymentContainer `json:"containers"`
}

type ocDeploymentTemplate struct {
	Spec ocDeploymentTemplateSpec `json:"spec"`
}

type ocDeploymentSpec struct {
	Template ocDeploymentTemplate `json:"template"`
}
type ocDeployment struct {
	Spec ocDeploymentSpec `json:"spec"`
}

var (
	consoleCmdArgs struct {
		workspaceContainerName string
		workspaceContainerPort string
	}
)

// openshiftConsoleCmd represents the openshiftConsole command
var openshiftConsoleCmd = &cobra.Command{
	Use:   "openshiftConsole",
	Short: "Launches an OpenShift console.",
	Long:  `Launches an OpenShift console application in a separate container.`,
	Run: func(cmd *cobra.Command, args []string) {
		ocUser := viper.GetString("ocUser")
		userHome := viper.GetString("userHome")
		out, err := pkgIntHelper.RunCommandPipeStdin("ocm", "post", "/api/accounts_mgmt/v1/access_token")
		if err != nil {
			logger.Fatal("Failed to run command: ", err)
		}
		path := fmt.Sprintf("%s/.kube/ocm-pull-secret/config.json", userHome)
		file, err := os.OpenFile(path, os.O_WRONLY, 0644)
		if err != nil {
			logger.Fatal("Failed to open file: ", err)
		}
		defer file.Close()
		file.WriteString(string(out))

		containerName := fmt.Sprintf("%s-openshift-console", consoleCmdArgs.workspaceContainerName)
		kubeConfigFileName := fmt.Sprintf("%s/.kube/ocm-pull-secret/config.json", userHome)
		consoleListenAddr := fmt.Sprintf("http://0.0.0.0:%s", consoleCmdArgs.workspaceContainerPort)
		pullArgs := []string{"pull", "--quiet", "--authfile", kubeConfigFileName}
		runArgs := []string{
			"run",
			"--rm",
			"--network",
			fmt.Sprintf("container:%s", consoleCmdArgs.workspaceContainerName),
			"-e",
			"HTTPS_PROXY=http://squid.corp.redhat.com:3128",
			"--name",
			containerName,
			"--authfile",
			kubeConfigFileName,
		}

		out, err = pkgIntHelper.RunCommandOutput(
			"podman",
			"exec",
			"-it",
			"--user",
			ocUser,
			consoleCmdArgs.workspaceContainerName,
			"oc",
			"get",
			"deployment",
			"console",
			"-n",
			"openshift-console",
			"-o",
			"json",
		)
		if err != nil {
			logger.Fatal("Failed to run command: ", err)
		}
		var openShiftConsoleDeploy ocDeployment
		var consoleImage string
		err = json.Unmarshal(out, &openShiftConsoleDeploy)
		if err != nil {
			logger.Fatal("Failed to unmarshal: ", err)
		}
		fmt.Println(openShiftConsoleDeploy)

		for _, container := range openShiftConsoleDeploy.Spec.Template.Spec.Containers {
			if container.Name == "console" {
				consoleImage = container.Image
				break
			}
		}

		out, err = pkgIntHelper.RunCommandOutput(
			"podman",
			"exec",
			"-it",
			"--user",
			ocUser,
			consoleCmdArgs.workspaceContainerName,
			"oc",
			"config",
			"view",
			"-o",
			"json",
		)
		if err != nil {
			logger.Fatal("Failed to run command: ", err)
		}

		var config pkgIntHelper.OcConfig
		err = json.Unmarshal(out, &config)
		if err != nil {
			logger.Fatal("Failed to unmarshal: ", err)
		}

		imagePullArgs := append(pullArgs, consoleImage)
		_, err = pkgIntHelper.RunCommandOutput(
			"podman",
			imagePullArgs...,
		)
		if err != nil {
			logger.Fatal("Failed to run command: ", err)
		}
		cluster := config.Clusters[0]
		apiUrl := cluster.ClusterUrls.Server
		alertManagerUrl := strings.Replace(apiUrl, "/backplane/cluster", "/backplane/alertmanager", 1)
		thanosUrl := strings.Replace(apiUrl, "/backplane/cluster", "/backplane/thanos", 1)
		alertManagerUrl = strings.TrimRight(alertManagerUrl, "/")
		thanosUrl = strings.TrimRight(thanosUrl, "/")

		out, err = pkgIntHelper.RunCommandOutput(
			"podman",
			"exec",
			"-it",
			"--user",
			ocUser,
			consoleCmdArgs.workspaceContainerName,
			"ocm",
			"token",
		)
		if err != nil {
			logger.Fatal("Failed to run command: ", err)
		}
		ocmToken := strings.TrimSpace(string(out))
		baseAddress := fmt.Sprintf("http://127.0.0.1:%s", consoleCmdArgs.workspaceContainerPort)
		runArgs = append(
			runArgs,
			consoleImage,
			"/opt/bridge/bin/bridge",
			"--public-dir",
			"/opt/bridge/static",
			"-base-address",
			baseAddress,
			"-branding",
			"dedicated",
			"-documentation-base-url",
			"https://docs.openshift.com/dedicated/4/",
			"-user-settings-location",
			"localstorage",
			"-user-auth",
			"disabled",
			"-k8s-mode",
			"off-cluster",
			"-k8s-auth",
			"bearer-token",
			"-k8s-mode-off-cluster-endpoint",
			apiUrl,
			"-k8s-mode-off-cluster-alertmanager",
			alertManagerUrl,
			"-k8s-mode-off-cluster-thanos",
			thanosUrl,
			"-k8s-auth-bearer-token",
			ocmToken,
			"-listen",
			consoleListenAddr,
			"-v",
			"5",
		)
		pkgIntHelper.RunCommandWithOsFiles("podman", os.Stdout, os.Stderr, os.Stdin, runArgs...)
	},
}

func init() {
	rootCmd.AddCommand(openshiftConsoleCmd)

	flags := openshiftConsoleCmd.Flags()
	flags.StringVarP(
		&consoleCmdArgs.workspaceContainerName,
		"workspaceContainerName",
		"c",
		"",
		"The running workspace container name that is logged into an OpenShift cluster.",
	)

	flags.StringVarP(
		&consoleCmdArgs.workspaceContainerPort,
		"workspaceContainerPort",
		"p",
		"",
		"The running workspace container port that is logged into an OpenShift cluster.",
	)

	openshiftConsoleCmd.MarkFlagRequired("workspaceContainerName")
	openshiftConsoleCmd.MarkFlagRequired("workspaceContainerPort")
}
