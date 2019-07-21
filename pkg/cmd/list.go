// Copyright 2019 Tetrate
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
	"os"

	"github.com/spf13/cobra"
	"github.com/tetratelabs/getenvoy/pkg/manifest"
)

// ListCmd lists available Envoy binaries
var (
	manifestURL string
	ListCmd     = &cobra.Command{
		Use:   "list",
		Short: "Lists available Envoys from GetEnvoy.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return manifest.Print(os.Stdout, manifestURL)
		},
	}
)

func init() {
	ListCmd.Flags().StringVarP(&manifestURL, "url", "u",
		"https://bintray.com/tetrate/getenvoy/download_file?file_path=manifest.json", "overrides the manifest URL")
	rootCmd.AddCommand(ListCmd)
}
