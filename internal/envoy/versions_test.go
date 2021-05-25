// Copyright 2021 Tetrate
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

package envoy

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/getenvoy/internal/version"
)

func TestPrintVersions(t *testing.T) {
	tests := []struct {
		name, platform string
		versions       version.EnvoyVersions
		expected       string
	}{
		{
			name:     "darwin",
			platform: "darwin/amd64",
			versions: goodVersions,
			expected: `1.18.3
1.14.6
`,
		},
		{
			name:     "linux",
			platform: "linux/amd64",
			versions: goodVersions,
			expected: `1.17.3
1.14.6
`,
		},
		{
			name:     "unsupported OS",
			platform: "windows/amd64",
			versions: goodVersions,
		},
		{
			name:     "unsupported Arch",
			platform: "linux/arm64",
			versions: goodVersions,
		},
		{
			name:     "empty version list",
			platform: "darwin/amd64",
			expected: ``,
		},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			out := new(bytes.Buffer)
			PrintVersions(tc.versions, tc.platform, out)
			require.Equal(t, tc.expected, out.String())
		})
	}
}

var goodVersions = version.EnvoyVersions{
	LatestVersion: "1.18.3",
	Versions: map[string]version.EnvoyVersion{
		"1.14.6": {
			Tarballs: map[string]string{
				"linux/amd64":  "https://getenvoy.io/versions/1.14.6/envoy-1.14.6-linux-x86_64.tar.gz",
				"darwin/amd64": "https://getenvoy.io/versions/1.14.6/envoy-1.14.6-darwin/amd64-x86_64.tar.gz",
			},
		},
		"1.17.3": {
			Tarballs: map[string]string{
				"linux/amd64": "https://getenvoy.io/versions/1.17.3/envoy-1.17.3-linux-x86_64.tar.gz",
			},
		},
		"1.18.3": {
			Tarballs: map[string]string{
				"darwin/amd64": "https://getenvoy.io/versions/1.18.3/envoy-1.18.3-darwin/amd64-x86_64.tar.gz",
			},
		},
	},
}