// Copyright © 2020 NAME HERE <EMAIL ADDRESS>
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
	"fmt"
	"github.com/aquasecurity/trivy/pkg/app"
	"github.com/spf13/cobra"
	"github.com/urfave/cli"
)

// scanImageCmd represents the scanImage command
var scanImageCmd = &cobra.Command{
	Use:   "scan-image",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		trivyApp := app.NewApp("dev")

		fmt.Println(trivyApp.Version)

		//test := []string{"--client", " --remote", "alpine:3.10"}
			//" --remote http://localhost:8080", "alpine:3.10"}
		/*cliargs := []string{
			"client",
			" --remote",
			"http://localhost:8081",
			"alpine:3.10",
		}*/

		cCmd := app.NewClientCommand()
		err := cCmd.Run(&cli.Context{})

		trivyApp.Commands[0].Run()

		//err := trivyApp.Run(cCmd)


		if err != nil {
			log.Fatal(err)
		}

	},
}

func init() {
	rootCmd.AddCommand(scanImageCmd)
	scanImageCmd.Flags().StringVarP(&imageName, "image", "i", "", "")
	scanImageCmd.MarkFlagRequired("image")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// scanImageCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// scanImageCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
