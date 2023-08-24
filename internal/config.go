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

package internal

import (
	"log"

	"github.com/spf13/viper"
)

type DirMap struct {
	HostDir      string `mapstructure:"hostDir"`
	ContainerDir string `mapstructure:"containerDir"`
	FileAttrs    string `mapstructure:"fileAttrs"`
}

type PortMap struct {
	HostPort      string `mapstructure:"hostPort"`
	ContainerPort string `mapstructure:"containerPort"`
}

type Plugin struct {
	Name          string `mapstructure:"name"`
	ExecPath      string `mapstructure:"execPath"`
	Config        string `mapstructure:"config"`
	RunOn         string `mapstructure:"runOn"`
	AllocatePorts int    `mapstructure:"allocatePorts"`
	ExecCommand   string `mapstructure:"execCommand"`
}

type OcmWorkspaceConfig struct {
	CustomDirMaps         []DirMap  `mapstructure:"customDirMaps"`
	AddToPATHEnv          []string  `mapstructure:"addToPATHEnv"`
	ExportEnvVars         []string  `mapstructure:"exportEnvVars"`
	HostUser              string    `mapstructure:"hostUser"`
	OcUser                string    `mapstructure:"ocUser"`
	UserHome              string    `mapstructure:"userHome"`
	BackplaneConfigProd   string    `mapstructure:"backplaneConfigProd"`
	BackplaneConfigStage  string    `mapstructure:"backplaneConfigStage"`
	BaseImage             string    `mapstructure:"baseImage"`
	OCMCLIVersion         string    `mapstructure:"ocmCLIVersion"`
	BackplaneCLIVersion   string    `mapstructure:"backplaneCLIVersion"`
	Plugins               []Plugin  `mapstructure:"plugins"`
	CustomPortMaps        []PortMap `mapstructure:"customPortMaps"`
	OcmLongLivedTokenPath string    `mapstructure:"ocmLongLivedTokenPath"`
}

func NewOcmWorkspaceConfig() *OcmWorkspaceConfig {
	var conf OcmWorkspaceConfig
	err := viper.Unmarshal(&conf)
	if err != nil {
		log.Fatal("Failed to unmarshal config: ", err)
	}
	return &conf
}

func (c *OcmWorkspaceConfig) GetAddToPATHEnv() []string {
	return c.AddToPATHEnv
}

func (c *OcmWorkspaceConfig) GetCustomDirMaps() []DirMap {
	return c.CustomDirMaps
}

func (c *OcmWorkspaceConfig) GetExportEnvVars() []string {
	return c.ExportEnvVars
}

func (c *OcmWorkspaceConfig) GetHostUser() string {
	return c.HostUser
}

func (c *OcmWorkspaceConfig) GetOcUser() string {
	return c.OcUser
}

func (c *OcmWorkspaceConfig) GetUserHome() string {
	return c.UserHome
}

func (c *OcmWorkspaceConfig) GetBackplaneConfigProd() string {
	return c.BackplaneConfigProd
}

func (c *OcmWorkspaceConfig) GetBackplaneConfigStage() string {
	return c.BackplaneConfigStage
}

func (c *OcmWorkspaceConfig) GetBaseImage() string {
	return c.BaseImage
}

func (c *OcmWorkspaceConfig) GetOcmCLIVersion() string {
	return c.OCMCLIVersion
}

func (c *OcmWorkspaceConfig) GetBackplaneCLIVersion() string {
	return c.BackplaneCLIVersion
}

func (c *OcmWorkspaceConfig) GetPlugins() []Plugin {
	return c.Plugins
}

func (c *OcmWorkspaceConfig) GetOcmLongLivedTokenPath() string {
	return c.OcmLongLivedTokenPath
}
