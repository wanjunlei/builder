#!/usr/bin/env bash
# Copyright 2020 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# The build.sh script builds stack images for the gcp/base builder.
#
# The script builds the following two images:
#   gcr.io/buildpacks/gcp/run:$tag
#   gcr.io/buildpacks/gcp/build:$tag
#
# It also validates that the build image includes all required licenses.
#
# Usage:
#   ./build.sh <path to self> <path to licenses.tar>

set -euo pipefail

# Convenient way to find the runfiles directory containing the Dockerfiles.
# $0 is in a different directory top-level directory.
DIR="$(dirname "$1")"
LICENSES="$2"

TAG="v2"

# Extract licenses.tar because it is symlinked, which Docker does not support.
readonly TEMP="$(mktemp -d)"
trap "rm -rf $TEMP" EXIT

echo "> Extracting licenses tar"
mkdir -p "$TEMP/licenses"
tar xf "$LICENSES" -C "$TEMP/licenses"

framework_version=< FRAMEWORK_VERSION >
if [[ "$framework_version" == *SNAPSHOT* ]]; then
  snapshotUrl="https://s01.oss.sonatype.org/content/repositories/snapshots/dev/openfunction/functions/functions-framework-invoker/$framework_version/"
  timestamp=$(curl -s ${snapshotUrl}maven-metadata.xml | grep -oPm1 "(?<=<timestamp>)[^<]+")
  buildNumber=$(curl -s ${snapshotUrl}maven-metadata.xml | grep -oPm1 "(?<=<buildNumber>)[^<]+")
  version=$(echo "${framework_version}" | sed s/-SNAPSHOT//)
  framework_url="${snapshotUrl}/functions-framework-invoker-$version-$timestamp-$buildNumber-jar-with-dependencies.jar"
else
  framework_url="https://repo.maven.apache.org/maven2/dev/openfunction/functions/functions-framework-invoker/$framework_version/functions-framework-invoker-$framework_version-jar-with-dependencies.jar"
fi

echo "> Building base run image"
docker build --build-arg "from_image=ibm-semeru-runtimes:open-< JAVA_VERSION >-jre" -t "java< JAVA_VERSION >-run" - < "${DIR}/parent.Dockerfile"
echo "> Building base build image"
docker build --build-arg "from_image=ibm-semeru-runtimes:open-< JAVA_VERSION >-jdk" -t "java< JAVA_VERSION >-build" - < "${DIR}/parent.Dockerfile"
echo "> Building run imager"
docker build --build-arg "from_image=java< JAVA_VERSION >-run" --build-arg "framework_url=${framework_url}" -t "< REGISTRY >/buildpacks-java< JAVA_VERSION >-run:$TAG" - < "${DIR}/run.Dockerfile"
#docker push < REGISTRY >/buildpacks-java< JAVA_VERSION >-run:$TAG
echo "> Building build imager"
docker build --build-arg "from_image=java< JAVA_VERSION >-build" -t "< REGISTRY >/buildpacks-java< JAVA_VERSION >-build:$TAG" -f "${DIR}/build.Dockerfile" "${TEMP}"
#docker push < REGISTRY >/buildpacks-java< JAVA_VERSION >-build:$TAG