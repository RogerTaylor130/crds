
set -o errexit
set -o nounset
set -o pipefail


SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
echo $SCRIPT_ROOT
CODEGEN_PKG=${CODEGEN_PKG:-$(cd "${SCRIPT_ROOT}"; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}
echo $CODEGEN_PKG