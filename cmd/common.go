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
	"errors"
	"fmt"
	"os"
	"strings"

	logger "github.com/sirupsen/logrus"

	pkgInt "ocm-workspace/internal"
)

type containerEngine interface {
	appendEnvVar(key string, value string)
	getEnvVars() [][]string
	appendVolMap(hostVol string, containerVol string, mapAttrs string)
	getVolMaps() [][]string
	appendPortMap(hostPort string, containerPort string, hostAddr string)
	getPortMaps() [][]string
	toEnvVarArgs() []string
	toVolMapArgs() []string
	toPortMapArgs() []string
	getRunArgs(containerName string, entryPoint string, image string, entryPointArgs ...string) []string
	getExecName() string
	appendBuildArg(name string, value string)
	toBuildArgs() []string
	getBuildArgs() []string
}

type containerEngineFactory interface {
	create() (containerEngine, error)
}

type ocmWorkspaceContainer struct {
	HostUser         string
	UserHome         string
	IsOcmLoginOnly   string
	CUSTOM_PORT_MAPS string
	UserBashrcPath   string
	OcmCluster       string
	OcmToken         string
	OcmEnvironment   string
}

func NewOcmWorkspaceContainer(config *pkgInt.OcmWorkspaceConfig) *ocmWorkspaceContainer {
	if config == nil {
		logger.Fatal("Config is not yet available at this stage.")
	}
	return &ocmWorkspaceContainer{
		HostUser:         getEnvVar("HOST_USER"),
		UserHome:         config.UserHome,
		IsOcmLoginOnly:   getEnvVar("IS_OCM_LOGIN_ONLY"),
		CUSTOM_PORT_MAPS: getEnvVar("CUSTOM_PORT_MAPS"),
		UserBashrcPath:   fmt.Sprintf("%s/.bashrc", config.UserHome),
		OcmCluster:       getEnvVar("OCM_CLUSTER"),
		OcmToken:         getEnvVar("OCM_TOKEN"),
		OcmEnvironment:   getEnvVar("OCM_ENVIRONMENT"),
	}
}

func checkContainerCommand() error {
	if !isInContainer() {
		return errors.New("this command is intended to be run only inside the workspace container")
	}
	return nil
}

func isInContainer() bool {
	return os.Getenv("IS_IN_CONTAINER") == "true"
}

func getEnvVar(name string) string {
	return strings.TrimSpace(os.Getenv(name))
}

func createConfigFile(path string, content string) {
	fayl, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		logger.Fatalf("Failed to open file %s: %v", path, err)
	}
	defer fayl.Close()
	fayl.WriteString(content)
}
