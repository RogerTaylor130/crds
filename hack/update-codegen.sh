#!/usr/bin/env bash

# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail


#function kube::codegen::internal::grep() {
#    # We use `grep` rather than `git grep` because sometimes external projects
#    # use this across repos.
#    grep "$@" \
#        --exclude-dir .git \
#        --exclude-dir _output \
#        --exclude-dir vendor
#}


SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
echo $SCRIPT_ROOT
CODEGEN_PKG=${CODEGEN_PKG:-$(cd "${SCRIPT_ROOT}"; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}
echo $CODEGEN_PKG
source "${CODEGEN_PKG}/kube_codegen.sh"

#THIS_PKG_PRE="crds"
#
#while read -r dir; do
#    obj_and_version="$(echo ${dir} | awk -F '/' '{print $(NF-1)"/"$NF}' )"
#    obj="$(echo ${dir} | awk -F '/' '{print $(NF-1)}' )"
#    echo "obj_and_version: ${obj_and_version}. obj: ${obj}"
#    kube::codegen::gen_helpers \
#        --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
#        "${SCRIPT_ROOT}/pkg/apis/${obj}"
#
#    kube::codegen::gen_client \
#        --with-watch \
#        --output-dir "${SCRIPT_ROOT}/${obj_and_version}/pkg/generated" \
#        --output-pkg "${THIS_PKG_PRE}/${obj_and_version}/pkg/generated" \
#        --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
#        "${SCRIPT_ROOT}/pkg/apis/${obj}"
#done < <(
#    ( kube::codegen::internal::grep -l --null \
#        -e '^\s*//\s*+k8s:deepcopy-gen=' \
#        -r "${SCRIPT_ROOT}/pkg/apis" \
#        --include '*.go' \
#        || true \
#    ) | while read -r -d $'\0' F; do dirname "${F}"; done \
#      | LC_ALL=C sort -u
#)

kube::codegen::gen_helpers \
    --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
    "${SCRIPT_ROOT}/pkg/apis"

kube::codegen::gen_client \
    --with-watch \
    --output-dir "${SCRIPT_ROOT}/pkg/generated" \
    --output-pkg "${THIS_PKG}/pkg/generated" \
    --boilerplate "${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
    "${SCRIPT_ROOT}/pkg/apis"

