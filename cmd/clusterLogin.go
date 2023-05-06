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
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var (
	clusterLoginCmdArgs struct {
		cluster string
	}
)

var clusterLoginCmd = &cobra.Command{
	Use:   "clusterLogin",
	Short: "Logs in to an OpenShift Dedicated cluster.",
	Run: func(cmd *cobra.Command, args []string) {
		configureOcmUser()
		configureDirs()
		ocmLogin()
		ocmBackplaneLogin()
		initTerminal()
	},
	Args: cobra.ExactArgs(1),
}

func initTerminal() {
	ocmUser := strings.TrimSpace(os.Getenv("OCM_USER"))
	userHome := fmt.Sprintf("/home/%s", ocmUser)
	cluster := strings.TrimSpace(os.Getenv("OCM_CLUSTER"))
	userBashrcPath := fmt.Sprintf("%s/.bashrc", userHome)
	runCommandStreamOutput("cp", "/terminal/bashrc", userBashrcPath)

	file, err := os.OpenFile(userBashrcPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open %s: %s\n", userBashrcPath, err)
	} else {
		defer file.Close()

		ps1String := fmt.Sprintf(
			"\nPS1='[%s@%s $(/usr/bin/workspace currentNamespace %s)]$ '\n",
			ocmUser, cluster, ocmUser,
		)
		_, err = file.WriteString(ps1String)
		if err != nil {
			log.Printf("Failed to write to file %s: %s\n", userBashrcPath, err)
		}

	}
	cmd := exec.Command(
		"sudo",
		"-Eu",
		ocmUser,
		"bash",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Run()
}

func ocmBackplaneLogin() {
	ocmUser := strings.TrimSpace(os.Getenv("OCM_USER"))
	cluster := strings.TrimSpace(os.Getenv("OCM_CLUSTER"))
	status := runCommandStreamOutput(
		"sudo",
		"-Eu",
		ocmUser,
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

func ocmLogin() {
	ocmToken := strings.TrimSpace(os.Getenv("OCM_TOKEN"))
	ocmEnvironment := strings.TrimSpace(os.Getenv("OCM_ENVIRONMENT"))
	ocmUser := strings.TrimSpace(os.Getenv("OCM_USER"))

	log.Println("Logging into ocm", ocmEnvironment)

	status := runCommandStreamOutput(
		"sudo",
		"-Eu",
		ocmUser,
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

func configureDirs() {
	ocmUser := strings.TrimSpace(os.Getenv("OCM_USER"))
	userHome := fmt.Sprintf("/home/%s", ocmUser)
	commands := [][]string{
		{
			"mkdir",
			"-p",
			fmt.Sprintf("%s/.kube", userHome),
		},
		{
			"chown",
			"-R",
			fmt.Sprintf("%s:%s", ocmUser, ocmUser),
			fmt.Sprintf("%s/.kube", userHome),
		},
		{
			"mkdir",
			"-p",
			fmt.Sprintf("%s/.config/ocm", userHome),
		},
		{
			"mkdir",
			"-p",
			fmt.Sprintf("%s/.config/backplane", userHome),
		},
		{
			"cp",
			"/backplane-config",
			fmt.Sprintf("%s/.config/backplane/", userHome),
		},
		{
			"chown",
			"-R",
			fmt.Sprintf("%s:%s", ocmUser, ocmUser),
			fmt.Sprintf("%s/.config/ocm", userHome),
		},
	}
	runCommandListStreamOutput(commands)
}

func configureOcmUser() {
	ocmUser := strings.TrimSpace(os.Getenv("OCM_USER"))
	userHome := fmt.Sprintf("/home/%s", ocmUser)
	commands := [][]string{
		{
			"useradd",
			"-m",
			ocmUser,
			"-d",
			userHome,
		},
		{
			"usermod",
			"-aG",
			"wheel",
			ocmUser,
		},
	}
	runCommandListStreamOutput(commands)
	line := []byte("\n%wheel         ALL = (ALL) NOPASSWD: ALL\n")
	os.WriteFile("/etc/sudoer", line, 0644)
}

func init() {
	rootCmd.AddCommand(clusterLoginCmd)

	flags := clusterLoginCmd.Flags()
	flags.StringVar(
		&clusterLoginCmdArgs.cluster,
		"cluster",
		"",
		"Cluster name or id.",
	)
}
