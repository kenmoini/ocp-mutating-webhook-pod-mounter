#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

## Get Token Secret name
SECRET=$(oc get serviceaccounts default -o jsonpath='{.secrets[*].name}' | tr ' ' '\n' | grep 'token')

## Get Token CA certificate
CA_BUNDLE=$(oc get secret ${SECRET} -o jsonpath="{.data.ca\.crt}")

export CA_BUNDLE

if command -v envsubst >/dev/null 2>&1; then
    envsubst
else
    sed -e "s|\${CA_BUNDLE}|${CA_BUNDLE}|g"
fi