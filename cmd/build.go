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
	pkgIntHelper "ocm-workspace/internal/helpers"

	pkgInt "ocm-workspace/internal"

	logger "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds the OCM workspace image locally.",
	Long:  `Builds the OCM workspace image locally and tags it with ocm-workspace:latest:`,
	Run: func(cmd *cobra.Command, args []string) {

		ceFactory := pkgInt.NewCeFactory(map[string]interface{}{
			"ceName": "podman",
		})

		ce, err := ceFactory.Create()
		if err != nil {
			logger.Fatal("Failed to create container engine: ", err)
		}

		ce.AppendBuildArg("BASE_IMAGE", config.BaseImage)
		ce.AppendBuildArg("OCM_CLI_VERSION", config.OCMCLIVersion)
		ce.AppendBuildArg("BACKPLANE_CLI_VERSION", config.BackplaneCLIVersion)

		out, err := pkgIntHelper.RunCommandOutput(
			"git",
			"rev-parse",
			"HEAD",
		)
		if err != nil {
			logger.Fatal(err)
		}
		headSha := string(out)
		ce.AppendBuildArg("BUILD_SHA", headSha)

		buildArgs := ce.GetBuildArgs()
		pkgIntHelper.RunCommandStreamOutput(ce.GetExecName(), buildArgs...)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
