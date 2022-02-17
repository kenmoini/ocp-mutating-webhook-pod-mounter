# OpenShift MutatingWebHook Pod Mounter

This repository provides the needed constructs for a MutatingWebHook that will automatically attach ConfigMaps to Pods.

## Deploy

1. Create a namespace `sidecar-injector` in which the sidecar injector webhook is deployed:

```bash
oc new-project sidecar-injector
```

2. Create a signed cert/key pair and store it in a Kubernetes `secret` that will be consumed by sidecar injector deployment:

```bash
./scripts/generate-certificates.sh \
  --service sidecar-injector-webhook-svc \
  --secret sidecar-injector-webhook-certs \
  --namespace sidecar-injector
```

3. Patch the `MutatingWebhookConfiguration` by set `caBundle` with correct value from Kubernetes cluster:

```bash
cat openshift/mutatingwebhook.yml | \
  scripts/patch-ca-bundle.sh > \
  openshift/mutatingwebhook-ca-bundle.yml
```

4. Deploy resources:

```bash
oc create -n sidecar-injector -f openshift/nginx-configmap.yml
oc create -n sidecar-injector -f openshift/configmap-webhook-config.yml
oc create -n sidecar-injector -f openshift/deployment-webhook-svc.yml
oc create -n sidecar-injector -f openshift/service-webhook-svc.yml
oc create -n sidecar-injector -f openshift/mutatingwebhook-ca-bundle.yml
```

## Test & Verify

1. The sidecar inject webhook should be in running state:

```
oc -n sidecar-injector get pod
NAME                                                   READY   STATUS    RESTARTS   AGE
sidecar-injector-webhook-deployment-7c8bc5f4c9-28c84   1/1     Running   0          30s

oc -n sidecar-injector get deploy
NAME                                  READY   UP-TO-DATE   AVAILABLE   AGE
sidecar-injector-webhook-deployment   1/1     1            1           67s
```

2. Create new namespace `injection` and label it with `sidecar-injector=enabled`:

```
oc new-project injection

oc label namespace injection sidecar-injection=enabled

oc get namespace -L sidecar-injection
NAME                 STATUS   AGE   SIDECAR-INJECTION
default              Active   26m
injection            Active   13s   enabled
kube-node-lease      Active   26m
kube-public          Active   26m
kube-system          Active   26m
...
sidecar-injector     Active   17m
```

3. Deploy an app in the OpenShift cluster, take `busybox` app as an example

```
oc run busybox --image=busybox --restart=Never -n injection --overrides='{"apiVersion":"v1","metadata":{"annotations":{"sidecar-injector-webhook.kenmoini.me/inject":"yes"}}}' --command -- sleep infinity
```

4. Verify sidecar container is injected:

```
oc get pod
NAME                     READY     STATUS        RESTARTS   AGE
busybox                   2/2       Running       0          1m

oc -n injection get pod busybox -o jsonpath="{.spec.containers[*].name}"
busybox sidecar-nginx
```

## Cleanup/Uninstallation

```bash
oc delete namespace sidecar-injector
oc delete namespace injection

oc delete MutatingWebhookConfiguration/sidecar-injector-webhook-cfg
```