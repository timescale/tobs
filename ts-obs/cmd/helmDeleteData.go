/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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

// helmDeleteDataCmd represents the helm delete-data command
var helmDeleteDataCmd = &cobra.Command{
	Use:   "delete-data",
	Short: "A brief description of your command",
	Args:  cobra.ExactArgs(0),
	RunE:  helmDeleteData,
}

func init() {
	helmCmd.AddCommand(helmDeleteDataCmd)
}

func helmDeleteData(cmd *cobra.Command, args []string) error {
	var err error

	var name string
	name, err = cmd.Flags().GetString("name")
	if err != nil {
		return fmt.Errorf("could not uninstall Timescale Observability: %w", err)
	}

	var namespace string
	namespace, err = cmd.Flags().GetString("namespace")
	if err != nil {
		return fmt.Errorf("could not uninstall Timescale Observability: %w", err)
	}

	fmt.Println("Getting Persistent Volume Claims")
	pvcnames, err := KubeGetPVCNames(namespace, map[string]string{"release": name})
	if err != nil {
		return fmt.Errorf("could not uninstall Timescale Observability: %w", err)
	}

	fmt.Println("Removing Persistent Volume Claims")
	for _, s := range pvcnames {
		err = KubeDeletePVC(namespace, s)
		if err != nil {
			return fmt.Errorf("could not uninstall Timescale Observability: %w", err)
		}
	}

	return nil
}
