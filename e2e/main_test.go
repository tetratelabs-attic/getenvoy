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

package e2e

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/tetratelabs/getenvoy/internal/moreos"
)

//nolint:golint
const (
	getenvoyBinaryEnvKey   = "E2E_GETENVOY_BINARY"
	envoyVersionsURLEnvKey = "ENVOY_VERSIONS_URL"
	envoyVersionsJSON      = "envoy-versions.json"
	runTimeout             = 2 * time.Minute
)

var (
	getEnvoyPath        = "getenvoy" // getEnvoyPath holds a path to a 'getenvoy' binary under test.
	expectedMockHeaders = map[string]string{"User-Agent": "getenvoy/dev"}
)

// TestMain ensures the "getenvoy" binary is valid.
func TestMain(m *testing.M) {
	// As this is an e2e test, we execute all tests with a binary compiled earlier.
	path, err := readGetEnvoyPath()
	if err != nil {
		exitOnInvalidBinary(err)
	}
	getEnvoyPath = path

	versionLine, _, err := getEnvoyExec("--version")
	if err != nil {
		exitOnInvalidBinary(err)
	}

	// Allow local file override when a SNAPSHOT version
	if _, err := os.Stat(envoyVersionsJSON); err == nil && strings.Contains(versionLine, "SNAPSHOT") {
		s, err := mockEnvoyVersionsServer() // no defer s.Close() because os.Exit() subverts it
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to serve %s: %v\n", envoyVersionsJSON, err)
			os.Exit(1)
		}
		os.Setenv(envoyVersionsURLEnvKey, s.URL)
	}
	os.Exit(m.Run())
}

func exitOnInvalidBinary(err error) {
	fmt.Fprintf(os.Stderr, `failed to start e2e tests due to an invalid "getenvoy" binary: %v\n`, err)
	os.Exit(1)
}

// mockEnvoyVersionsServer ensures envoyVersionsURLEnvKey is set appropriately, so that non-release versions can see
// changes to local envoyVersionsJSON.
func mockEnvoyVersionsServer() (*httptest.Server, error) {
	f, err := os.Open(envoyVersionsJSON)
	if err != nil {
		return nil, err
	}

	defer f.Close() // nolint
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ensure e2e tests won't eventually interfere with analytics when run against a release version
		for k, v := range expectedMockHeaders {
			h := r.Header.Get(k)
			if h != v {
				w.WriteHeader(500)
				w.Write([]byte(fmt.Sprintf("invalid %q: %s != %s\n", k, h, v))) //nolint
				return
			}
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(b) //nolint
	}))
	return ts, nil
}

// readGetEnvoyPath reads E2E_GETENVOY_BINARY or defaults to "$PWD/dist/getenvoy_$GOOS_$GOARCH/getenvoy"
// An error is returned if the value isn't an executable file.
func readGetEnvoyPath() (string, error) {
	path := os.Getenv(getenvoyBinaryEnvKey)
	if path == "" {
		// Assemble the default created by "make bin"
		relativePath := filepath.Join("..", "dist", fmt.Sprintf("getenvoy_%s_%s", runtime.GOOS, runtime.GOARCH), "getenvoy")
		abs, err := filepath.Abs(relativePath)
		if err != nil {
			return "", fmt.Errorf("%s didn't resolve to a valid path. Correct environment variable %s", path, getenvoyBinaryEnvKey)
		}
		path = abs
	}

	stat, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return "", fmt.Errorf("%s doesn't exist. Correct environment variable %s", path, getenvoyBinaryEnvKey)
	}
	if stat.IsDir() {
		return "", fmt.Errorf("%s is not a file. Correct environment variable %s", path, getenvoyBinaryEnvKey)
	}
	// While "make bin" should result in correct permissions, double-check as some tools lose them, such as
	// https://github.com/actions/upload-artifact#maintaining-file-permissions-and-case-sensitive-files
	if !moreos.IsExecutable(stat) {
		return "", fmt.Errorf("%s is not executable. Correct environment variable %s", path, getenvoyBinaryEnvKey)
	}
	return path, nil
}

type getEnvoy struct {
	cmd      *exec.Cmd
	runDir   string
	envoyPid int32
}

func newGetEnvoy(ctx context.Context, args ...string) *getEnvoy {
	cmd := exec.CommandContext(ctx, getEnvoyPath, args...)
	cmd.SysProcAttr = moreos.ProcessGroupAttr()
	return &getEnvoy{cmd: cmd}
}

func (b *getEnvoy) String() string {
	return strings.Join(b.cmd.Args, " ")
}

func getEnvoyExec(args ...string) (string, string, error) {
	g := newGetEnvoy(context.Background(), args...)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	g.cmd.Stdout = io.MultiWriter(os.Stdout, stdout) // we want to see full `getenvoy` output in the test log
	g.cmd.Stderr = io.MultiWriter(os.Stderr, stderr)
	err := g.cmd.Run()
	return stdout.String(), stderr.String(), err
}
