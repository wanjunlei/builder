// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Implements java/gradle buildpack.
// The gradle buildpack builds Gradle applications.
package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/GoogleCloudPlatform/buildpacks/pkg/devmode"
	"github.com/GoogleCloudPlatform/buildpacks/pkg/env"
	gcp "github.com/GoogleCloudPlatform/buildpacks/pkg/gcpbuildpack"
)

const (
	defaultGradleVersion = "7.4.2"
	defaultGradleURL     = "https://services.gradle.org/distributions/gradle-%s-bin.zip"
	gradlePath           = "/usr/local/gradle"
)

func main() {
	gcp.Main(detectFn, buildFn)
}

func detectFn(ctx *gcp.Context) (gcp.DetectResult, error) {
	if ctx.FileExists("build.gradle") {
		return gcp.OptInFileFound("build.gradle"), nil
	}
	if ctx.FileExists("build.gradle.kts") {
		return gcp.OptInFileFound("build.gradle.kts"), nil
	}
	return gcp.OptOut("neither build.gradle nor build.gradle.kts found"), nil
}

func buildFn(ctx *gcp.Context) error {
	var gradle string
	if ctx.FileExists("gradlew") {
		gradle = "./gradlew"
	} else {
		if needToInstallGradle() {
			if err := installGradle(ctx); err != nil {
				return fmt.Errorf("installing Gradle: %w", err)
			}
		}

		gradle = "gradle"
	}

	ctx.Exec([]string{gradle, "-v"}, gcp.WithUserAttribution)

	command := []string{gradle, "clean", "assemble", "-x", "test", "--build-cache"}

	if buildArgs := os.Getenv(env.BuildArgs); buildArgs != "" {
		if strings.Contains(buildArgs, "project-cache-dir") {
			ctx.Warnf("Detected project-cache-dir property set in GOOGLE_BUILD_ARGS. Dependency caching may not work properly.")
		}
		command = append(command, buildArgs)
	}

	if !ctx.Debug() && !devmode.Enabled(ctx) {
		command = append(command, "--quiet")
	}

	ctx.Exec(command, gcp.WithUserAttribution)
	return nil
}

func needToInstallGradle() bool {
	if version := os.Getenv(env.GradleVersion); version != "" && version != defaultGradleVersion {
		return true
	}

	if url := os.Getenv(env.GradleURL); url != "" {
		return true
	}

	return false
}

// installGradle installs Gradle and returns the path of the gradle binary
func installGradle(ctx *gcp.Context) error {
	var gradleURL string
	if url := os.Getenv(env.GradleURL); url != "" {
		gradleURL = url
	} else {
		gradleVersion := os.Getenv(env.MavenVersion)
		gradleURL = fmt.Sprintf(defaultGradleURL, gradleVersion)
	}

	// Download and install gradle in layer.
	ctx.Logf("Installing Gradle from %s", gradleURL)
	if code := ctx.HTTPStatus(gradleURL); code != http.StatusOK {
		return fmt.Errorf("gradle does not exist at %s (status %d)", gradleURL, code)
	}

	tmpDir := "/tmp/gradle/"
	gradleZip := filepath.Join(tmpDir, "gradle.zip")
	defer ctx.RemoveAll(tmpDir)
	// download gradle
	curl := fmt.Sprintf("curl --fail --show-error --silent --location --retry 3 %s --output %s", gradleURL, gradleZip)
	ctx.Exec([]string{"bash", "-c", curl}, gcp.WithUserAttribution)
	// unzip
	unzip := fmt.Sprintf("unzip -q %s -d %s", gradleZip, tmpDir)
	ctx.Exec([]string{"bash", "-c", unzip}, gcp.WithUserAttribution)
	// delete the zip file
	ctx.Exec([]string{"bash", "-c", fmt.Sprintf("rm -rf %s", gradleZip)}, gcp.WithUserAttribution)

	ctx.Exec([]string{"bash", "-c", fmt.Sprintf("mv %s* %s", tmpDir, gradlePath)}, gcp.WithUserTimingAttribution)

	// delete the old gradle
	command := fmt.Sprintf("rm -rf %s/current %s/gradle-%s", gradlePath, gradlePath, defaultGradleVersion)
	ctx.Exec([]string{"bash", "-c", command}, gcp.WithUserAttribution)

	// link the new maven to current
	command = fmt.Sprintf("ln -s $(find /usr/local/gradle/ -type d -name \"gradle*\") %s/current", gradlePath)
	ctx.Exec([]string{"bash", "-c", command}, gcp.WithUserAttribution)

	return nil
}
