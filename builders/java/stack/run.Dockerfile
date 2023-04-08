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

ARG framework_url
ARG framework_path=/usr/local/openfunction/

ADD ${framework_url} /

RUN mkdir ${framework_path} && \
  mv /functions-framework-invoker* ${framework_path} && \
  ln -s $(find ${framework_path} -name "functions-framework-invoker*") ${framework_path}functions-framework.jar && \
  chown -R ${cnb_uid}:${cnb_gid} ${framework_path}

ENV PORT 8080
USER cnb

