/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

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
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

// clusterLoginCmd represents the clusterLogin command
var clusterLoginCmd = &cobra.Command{
	Use:   "clusterLogin",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Running ocm-workspace container.")

		signalChan := make(chan os.Signal)
		signal.Notify(signalChan, os.Interrupt, syscall.SIGINT)
		go func() {
			<-signalChan
			log.Println("Caught signal exiting.")
			os.Exit(1)
		}()

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
	runCommandStreamOutput(
		"mkdir",
		"-p",
		fmt.Sprintf("%s/.kube", userHome),
	)
	runCommandStreamOutput(
		"chown",
		"-R",
		fmt.Sprintf("%s:%s", ocmUser, ocmUser),
		fmt.Sprintf("%s/.kube", userHome),
	)
	runCommandStreamOutput(
		"mkdir",
		"-p",
		fmt.Sprintf("%s/.config/ocm", userHome),
	)
	runCommandStreamOutput(
		"mkdir",
		"-p",
		fmt.Sprintf("%s/.config/backplane", userHome),
	)
	runCommandStreamOutput(
		"cp",
		"/backplane-config",
		fmt.Sprintf("%s/.config/backplane/", userHome),
	)
	runCommandStreamOutput(
		"chown",
		"-R",
		fmt.Sprintf("%s:%s", ocmUser, ocmUser),
		fmt.Sprintf("%s/.config/ocm", userHome),
	)
}

func configureOcmUser() {
	ocmUser := strings.TrimSpace(os.Getenv("OCM_USER"))
	userHome := fmt.Sprintf("/home/%s", ocmUser)
	runCommandStreamOutput(
		"useradd",
		"-m",
		ocmUser,
		"-d",
		userHome,
	)
	runCommandStreamOutput(
		"usermod",
		"-aG",
		"wheel",
		ocmUser,
	)
	line := []byte("\n%wheel         ALL = (ALL) NOPASSWD: ALL\n")
	os.WriteFile("/etc/sudoer", line, 0644)
}

func init() {
	rootCmd.AddCommand(clusterLoginCmd)

	flags := clusterLoginCmd.Flags()
	flags.StringVar(
		&args.cluster,
		"cluster",
		"",
		"Cluster name or id.",
	)
}
