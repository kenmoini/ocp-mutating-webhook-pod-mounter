#!/bin/bash

oc delete namespace sidecar-injector
oc delete MutatingWebhookConfiguration/sidecar-injector-webhook-cfg