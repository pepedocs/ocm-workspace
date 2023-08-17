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
)

var (
	currentNamespaceCmdArgs struct {
		ocUser string
	}
)

var currentNamespaceCmd = &cobra.Command{
	Use:   "currentNamespace",
	Short: "Shows OpenShift's current context namespace given an OpenShift user.",
	Run: func(cmd *cobra.Command, args []string) {
		if !checkContainerCommand() {
			return
		}
		namespace, err := ocGetCurrentNamespace(currentNamespaceCmdArgs.ocUser)
		if err != nil {
			fmt.Print("na")
		}
		fmt.Print(namespace)
	},
}

func init() {
	rootCmd.AddCommand(currentNamespaceCmd)
	currentNamespaceCmd.Flags().StringVarP(
		&currentNamespaceCmdArgs.ocUser,
		"ocUser",
		"u",
		"",
		"Run as OpenShift user.",
	)
	currentNamespaceCmd.MarkFlagRequired("ocUser")
}
