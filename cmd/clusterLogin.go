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
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var clusterLoginCmd = &cobra.Command{
	Use:   "clusterLogin",
	Short: "Logs in to an OpenShift Dedicated cluster.",
	Run: func(cmd *cobra.Command, args []string) {
		isOcmLoginOnly, err := strconv.ParseBool(os.Getenv("IS_OCM_LOGIN_ONLY"))
		if err != nil {
			glog.Warning("Failed to parse environment variable: ", err)
		}
		configureOcmUser()
		configureDirs()
		ocmLogin()
		if !isOcmLoginOnly {
			ocmBackplaneLogin()
			processOpenShiftServiceReference()
		}
		initTerminal()
	},
}

// Processes OpenShift service references described in the ocm workpsace config
func processOpenShiftServiceReference() {
	var svcHostPortBindList serviceHostPortBindList
	envVar := strings.TrimSpace(os.Getenv("HOST_PORT_MAPS"))
	envVar = strings.Trim(envVar, "\"")
	err := yaml.Unmarshal([]byte(envVar), &svcHostPortBindList)
	if err != nil {
		log.Fatalf("Failed to unmarshal environment variable value HOST_PORT_MAPS %s: %s", envVar, err)
	}

	if len(svcHostPortBindList.ServiceHostPortBinds) == 0 {
		log.Println("No service/host ports to bind.")
	}

	ocUser := strings.TrimSpace(os.Getenv("OCM_USER"))
	for _, portBind := range svcHostPortBindList.ServiceHostPortBinds {
		params := []string{
			"-Eu",
			ocUser,
			"oc",
			"port-forward",
			portBind.SourceName,
			"--address",
			"0.0.0.0",
			fmt.Sprintf("%s:%s", strconv.Itoa(portBind.HostPort), strconv.Itoa(portBind.ServicePort)),
			"-n",
			portBind.Namespace,
		}
		log.Printf("Forwarding %s/%s port to %s", portBind.ParentService, portBind.ServiceName, strconv.Itoa(portBind.HostPort))
		cmd := exec.Command("sudo", params...)
		err := cmd.Start()
		if err != nil {
			log.Printf("Failed to port forward %s: %s\n", strconv.Itoa(portBind.ServicePort), err)
		}

	}
}

// Initializes a bash terminal for ocm workspace
func initTerminal() {
	ocUser := strings.TrimSpace(os.Getenv("OCM_USER"))
	userHome := fmt.Sprintf("/home/%s", ocUser)
	cluster := strings.TrimSpace(os.Getenv("OCM_CLUSTER"))
	userBashrcPath := fmt.Sprintf("%s/.bashrc", userHome)
	runCommandStreamOutput("cp", "/terminal/bashrc", userBashrcPath)

	file, err := os.OpenFile(userBashrcPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open %s: %s\n", userBashrcPath, err)
	} else {
		defer file.Close()

		ps1String := fmt.Sprintf(
			"\nPS1='[%s %s $(/usr/bin/workspace --config /.ocm-workspace.yaml currentNamespace -u %s)]$ '\n",
			ocUser, cluster, ocUser,
		)
		_, err = file.WriteString(ps1String)
		if err != nil {
			log.Printf("Failed to write to file %s: %s\n", userBashrcPath, err)
		}

	}
	cmd := exec.Command(
		"sudo",
		"-Eu",
		ocUser,
		"bash",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
}

// Logs in a user in to an OCM cluster
func ocmBackplaneLogin() {
	ocUser := strings.TrimSpace(os.Getenv("OCM_USER"))
	cluster := strings.TrimSpace(os.Getenv("OCM_CLUSTER"))
	status := runCommandStreamOutput(
		"sudo",
		"-Eu",
		ocUser,
		"ocm",
		"backplane",
		"login",
		cluster,
	)

	if status.Exit != 0 {
		log.Fatal("OCM backplane login failed.")
	}
	log.Println("OCM backplane login successful.")
}

// Logs in a user in to OCM
func ocmLogin() {
	ocmToken := strings.TrimSpace(os.Getenv("OCM_TOKEN"))
	ocmEnvironment := strings.TrimSpace(os.Getenv("OCM_ENVIRONMENT"))
	ocUser := strings.TrimSpace(os.Getenv("OCM_USER"))

	log.Println("Logging into ocm", ocmEnvironment)

	status := runCommandStreamOutput(
		"sudo",
		"-Eu",
		ocUser,
		"ocm",
		"login",
		fmt.Sprintf("--token=%s", ocmToken),
		fmt.Sprintf("--url=%s", ocmEnvironment),
	)

	if status.Exit != 0 {
		log.Fatal("OCM Login failed")
	}
	log.Println("OCM Login successful.")
}

// Configures the required directories in the ocm workspace container
func configureDirs() {
	ocUser := strings.TrimSpace(os.Getenv("OCM_USER"))
	userHome := fmt.Sprintf("/home/%s", ocUser)
	commands := [][]string{
		{
			"mkdir",
			"-p",
			fmt.Sprintf("%s/.kube", userHome),
		},
		{
			"chown",
			"-R",
			fmt.Sprintf("%s:%s", ocUser, ocUser),
			fmt.Sprintf("%s/.kube", userHome),
		},
		{
			"mkdir",
			"-p",
			fmt.Sprintf("%s/.config/ocm", userHome),
		},
		{
			"chown",
			"-R",
			fmt.Sprintf("%s:%s", ocUser, ocUser),
			fmt.Sprintf("%s/.config/ocm", userHome),
		},
	}
	runCommandListStreamOutput(commands)
}

// Configures the OCM user in the ocm workspace container
func configureOcmUser() {
	ocUser := strings.TrimSpace(os.Getenv("OCM_USER"))
	userHome := fmt.Sprintf("/home/%s", ocUser)
	commands := [][]string{
		{
			"useradd",
			"-m",
			ocUser,
			"-d",
			userHome,
		},
		{
			"usermod",
			"-aG",
			"wheel",
			ocUser,
		},
	}
	runCommandListStreamOutput(commands)
	line := []byte("\n%wheel         ALL = (ALL) NOPASSWD: ALL\n")
	os.WriteFile("/etc/sudoer", line, 0644)
}

func init() {
	rootCmd.AddCommand(clusterLoginCmd)
}
