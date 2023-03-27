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

// Implements java/maven buildpack.
// The maven buildpack builds Maven applications.
package main

import (
	"fmt"
	"github.com/GoogleCloudPlatform/buildpacks/pkg/env"
	gcp "github.com/GoogleCloudPlatform/buildpacks/pkg/gcpbuildpack"
	"github.com/buildpacks/libcnb"
	"net/http"
	"os"
)

const (
	layerName                         = "skywalking"
	defaultSkywalkingJavaAgentVersion = "8.14.0"
	defaultSkywalkingJavaAgentURL     = "https://archive.apache.org/dist/skywalking/java-agent/%[1]s/apache-skywalking-java-agent-%[1]s.tgz"
	defaultSkywalkingJavaAgentPath    = "/usr/local/skywalking-agent/"
)

func main() {
	gcp.Main(detectFn, buildFn)
}

func detectFn(_ *gcp.Context) (gcp.DetectResult, error) {
	return gcp.OptInAlways(), nil
}

func buildFn(ctx *gcp.Context) error {
	layer := ctx.Layer(layerName)
	layer.Launch = true

	if disable, ok := os.LookupEnv(env.DisableSkywalking); ok && disable == "true" {
		ctx.Logf("Delete skywalking java agent")
		deleteDefaultAgent(ctx)
		return nil
	}

	if url, ok := os.LookupEnv(env.SkywalkingJavaAgentURL); ok && len(url) > 0 {
		if err := downloadAgent(ctx, layer, url); err != nil {
			return err
		}
		setJavaAgentArg(ctx, layer)
	} else if version, ok := os.LookupEnv(env.SkywalkingJavaAgentVersion); ok {
		if version != defaultSkywalkingJavaAgentVersion {
			if err := downloadAgent(ctx, layer, fmt.Sprintf(defaultSkywalkingJavaAgentURL, version)); err != nil {
				return err
			}
			setJavaAgentArg(ctx, layer)
		}
	} else {
		setJavaAgentArg(ctx, nil)
	}

	return nil
}

func setJavaAgentArg(ctx *gcp.Context, layer *libcnb.Layer) {
	path := defaultSkywalkingJavaAgentPath
	if layer != nil {
		path = layer.Path
	}
	ctx.Setenv(env.JavaAgentPath, path+"skywalking-agent.jar")
}

func deleteDefaultAgent(ctx *gcp.Context) {
	command := fmt.Sprintf("rm -rf %s", defaultSkywalkingJavaAgentPath)
	ctx.Exec([]string{"bash", "-c", command}, gcp.WithUserAttribution)
}

func downloadAgent(ctx *gcp.Context, layer *libcnb.Layer, url string) error {
	if code := ctx.HTTPStatus(url); code != http.StatusOK {
		return gcp.UserErrorf("Skywalking java agent does not exist at %s (status %d).", url, code)
	}
	command := fmt.Sprintf("curl --fail --show-error --silent --location --retry 3 %s | tar xz --directory %s", url, layer.Path)
	ctx.Exec([]string{"bash", "-c", command}, gcp.WithUserAttribution)

	deleteDefaultAgent(ctx)
	return nil
}
