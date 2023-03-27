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

ARG from_image
FROM ${from_image}

ARG cnb_uid=${CNB_USER_ID}
ARG cnb_gid=${CNB_GROUP_ID}
ARG skywalking_java_agent_version=8.14.0
ARG opentelemetry_java_agent_version=v1.23.0
ARG framework_url
ARG framework_path=/usr/local/openfunction/

ADD https://archive.apache.org/dist/skywalking/java-agent/${skywalking_java_agent_version}/apache-skywalking-java-agent-${skywalking_java_agent_version}.tgz /
ADD https://github.com/open-telemetry/opentelemetry-java-instrumentation/releases/download/${opentelemetry_java_agent_version}/opentelemetry-javaagent.jar /
ADD ${framework_url} /

RUN tar xzf apache-skywalking-java-agent-${skywalking_java_agent_version}.tgz -C /usr/local/ && \
  rm -rf apache-skywalking-java-agent-${skywalking_java_agent_version}.tgz && \
  chown -R ${cnb_uid}:${cnb_gid} /usr/local/skywalking-agent && \
  mkdir /usr/local/opentelemetry && \
  mv /opentelemetry-javaagent.jar /usr/local/opentelemetry/ && \
  chown -R ${cnb_uid}:${cnb_gid} /usr/local/opentelemetry/ && \
  mkdir ${framework_path} && \
  mv /functions-framework-invoker* ${framework_path} && \
  ln -s $(find ${framework_path} -name "functions-framework-invoker*") ${framework_path}functions-framework.jar && \
  chown -R ${cnb_uid}:${cnb_gid} ${framework_path}

ENV PORT 8080
USER cnb

