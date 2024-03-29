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
	logger "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

// buildInfoCmd represents the buildInfo command
var buildInfoCmd = &cobra.Command{
	Use:   "buildInfo",
	Short: "Prints the current ocm-workspace image build information.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := checkContainerCommand(); err != nil {
			return
		}
		logger.Infof("Build SHA: %s\n", getEnvVar("BUILD_SHA"))
	},
}

func init() {
	rootCmd.AddCommand(buildInfoCmd)
}
