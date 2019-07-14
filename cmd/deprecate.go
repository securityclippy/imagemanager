// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/securityclippy/imagemanager/pkg/manager"

	"github.com/spf13/cobra"
)

var imageName string

// deprecateCmd represents the deprecate command
var deprecateCmd = &cobra.Command{
	Use:   "deprecate",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		m := manager.NewManager("", "", "f977944a-e11c-4ef2-89f8-acfd3581ae08")
		err := m.Deprecate(imageName)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(deprecateCmd)
	deprecateCmd.Flags().StringVarP(&imageName, "image", "i", "", "image to deprecate")
	deprecateCmd.MarkFlagRequired("image")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deprecateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deprecateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
