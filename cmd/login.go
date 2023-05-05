/*
Copyright © 2023 Jose Cueto <pepedocs@gmail.com>

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
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	args struct {
		cluster        string
		ocmEnvironment string
	}
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Runs the ocm workspace container and logs in to a cluster.",
	Long: `Runs the ocm workspace container with the ocm workspace as its entrypoint.
           The entrypoint is supplied with the "clusterLogin" flag to proceed with
		   logging into a cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		ocmCluster := args[0]
		ocmEnvironment := "production"
		if len(args) > 1 {
			ocmEnvironment = args[1]
		}
		runOCMWorkspaceContainer(ocmCluster, ocmEnvironment)
	},
	Args: cobra.MaximumNArgs(2),
}

func runOCMWorkspaceContainer(ocmCluster string, ocmEnvironment string) {
	envVarOcmUser := fmt.Sprintf("OCM_USER=%s", viper.GetString("ocmUser"))
	envVarOcmToken := fmt.Sprintf("OCM_TOKEN=%s", viper.GetString("ocmToken"))
	envVarCluster := fmt.Sprintf("OCM_CLUSTER=%s", ocmCluster)

	userHome := fmt.Sprintf("/home/%s", viper.GetString("ocmUser"))

	containerBackplaneConfigPath := "/backplane-config"

	volMapBackplaneConfig := fmt.Sprintf("%s/.config/backplane/config.prod.json:%s:ro", userHome, containerBackplaneConfigPath)
	if ocmEnvironment == "staging" {
		volMapBackplaneConfig = fmt.Sprintf("%s/.config/backplane/config.stage.json:/backplane-config:ro", userHome)
	}

	envVarOcmEnvironment := fmt.Sprintf("OCM_ENVIRONMENT=%s", ocmEnvironment)
	envVarBackplaneConfig := fmt.Sprintf("BACKPLANE_CONFIG=%s", containerBackplaneConfigPath)
	cmd := exec.Command(
		"podman",
		"run",
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
		"-v",
		volMapBackplaneConfig,
		"--entrypoint",
		"./workspace",
		"ocm-workspace:latest",
		"clusterLogin",
		ocmCluster,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
}

func init() {
	rootCmd.AddCommand(loginCmd)

	flags := loginCmd.Flags()
	flags.StringVarP(
		&args.cluster,
		"ocmCluster",
		"c",
		"",
		"Cluster name or id.",
	)

	flags.StringVarP(
		&args.ocmEnvironment,
		"ocmEnvironment",
		"e",
		"production",
		"OCM environemnt (production, staging)",
	)
}
