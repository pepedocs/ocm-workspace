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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds the OCM workspace image locally.",
	Long:  `Builds the OCM workspace image locally and tags it with ocm-workspace:latest:`,
	Run: func(cmd *cobra.Command, args []string) {
		buildArgBaseImage := fmt.Sprintf("BASE_IMAGE=%s", viper.GetString("baseImage"))
		buildArgOcmCLIVersion := fmt.Sprintf("OCM_CLI_VERSION=%s", viper.GetString("ocmCLIVersion"))
		buildArgRhocCLIVersion := fmt.Sprintf("RHOC_CLI_VERSION=%s", viper.GetString("rhocCLIVersion"))
		buildArgBackplaneCLIVersion := fmt.Sprintf("BACKPLANE_CLI_VERSION=%s", viper.GetString("backplaneCLIVersion"))

		runCommandStreamOutput(
			"podman",
			"build",
			"-t",
			"ocm-workspace",
			"--build-arg",
			buildArgBaseImage,
			"--build-arg",
			buildArgOcmCLIVersion,
			"--build-arg",
			buildArgRhocCLIVersion,
			"--build-arg",
			buildArgBackplaneCLIVersion,
			".",
		)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
