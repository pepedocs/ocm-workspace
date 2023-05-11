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

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	consoleCmdArgs struct {
		workspaceContainerName string
		workspaceContainerPort int
	}
)

// openshiftConsoleCmd represents the openshiftConsole command
var openshiftConsoleCmd = &cobra.Command{
	Use:   "openshiftConsole",
	Short: "Launches an OpenShift console.",
	Long:  `Launches an OpenShift console application in a separate container.`,
	Run: func(cmd *cobra.Command, args []string) {
		userHome := fmt.Sprintf("/home/%s", viper.GetString("hostUser"))
		out, err := runCommandPipeStdin("ocm", "post", "/api/accounts_mgmt/v1/access_token")
		if err != nil {
			glog.Fatal("Failed to run command: ", err)
		}
		path := fmt.Sprintf("%s/.kube/ocm-pull-secret/config.json", userHome)
		file, err := os.OpenFile(path, os.O_WRONLY, 0644)
		if err != nil {
			glog.Fatal("Failed to open file: ", err)
		}
		defer file.Close()
		file.WriteString(string(out))
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

	port := fmt.Sprintf("%v", &consoleCmdArgs.workspaceContainerPort)
	flags.StringVarP(
		&port,
		"workspaceContainerPort",
		"p",
		"",
		"The running workspace container port that is logged into an OpenShift cluster.",
	)

	openshiftConsoleCmd.MarkFlagRequired("workspaceContainerName")
	openshiftConsoleCmd.MarkFlagRequired("workspaceContainerPort")
}
