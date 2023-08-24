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
	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"

	pkgInt "ocm-workspace/internal"
)

var config *pkgInt.OcmWorkspaceConfig
var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "workspace",
	Short: "A containerised workspace for managing OpenShift Dedicated",
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "verbose logging")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ocm-workspace.yaml)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		var src string
		if os.Getenv("IS_IN_CONTAINER") == "true" {
			src = "/"
		} else {
			home, err := os.UserHomeDir()
			src = home
			cobra.CheckErr(err)
		}
		viper.AddConfigPath(src)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".ocm-workspace")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err != nil {
		glog.Fatal("Failed to read config file: ", err)
	}

	config = pkgInt.NewOcmWorkspaceConfig()
}
