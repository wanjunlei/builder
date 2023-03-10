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
	defaultSkywalkingJavaAgentURL     = "https://dlcdn.apache.org/skywalking/java-agent/%[1]s/apache-skywalking-java-agent-%[1]s.tgz"
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
		path, ok := os.LookupEnv(env.SkywalkingJavaAgentPath)
		if !ok || len(path) == 0 {
			return gcp.Errorf(gcp.StatusNotFound, "%s must be set when install skywalking agentfrom url", env.SkywalkingJavaAgentPath)
		}

		if err := downloadAgent(ctx, layer, url); err != nil {
			return err
		}
		setJavaAgentArg(ctx, layer, path)
		return nil
	}

	if version, ok := os.LookupEnv(env.SkywalkingJavaAgentVersion); ok {
		if version != defaultSkywalkingJavaAgentVersion {
			if err := downloadAgent(ctx, layer, fmt.Sprintf(defaultSkywalkingJavaAgentURL, version)); err != nil {
				return err
			}
			setJavaAgentArg(ctx, layer, "")
			return nil
		}
	}

	setJavaAgentArg(ctx, nil, "")
	return nil
}

func setJavaAgentArg(ctx *gcp.Context, layer *libcnb.Layer, path string) {
	arg := "-javaagent:"
	absPath := defaultSkywalkingJavaAgentPath + "skywalking-agent.jar"
	if layer != nil {
		absPath = layer.Path
		if path == "" {
			absPath += "skywalking-agent/skywalking-agent.jar"
		} else {
			absPath += path
		}
	}
	ctx.Setenv(env.SkywalkingJavaAgentArg, arg+absPath)
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
