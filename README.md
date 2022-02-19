# DEPRECIATED - Use the [PodPreset Operator](https://github.com/redhat-cop/podpreset-webhook)

This project was intending to be just a MutatingWebhook that would auto-mount a ConfigMap to Pods that contained a Root CA Bundle PEM and would then also bake a Java Keystore ConfigMap and attach it as well.

Instead, you can simply create/sync the ConfigMaps then use the PodPreset Operator to auto-attach the Volumes and VolumeMounts pointing to the ConfigMaps.

You can create a ConfigMap file with the Java Keystore as binary data with the following:

```bash
oc create configmap root-jks --from-file=/etc/pki/ca-trust/extracted/java/cacerts -o yaml --dry-run=client > root-jks.yaml
```

And maybe a PodPreset like this:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: PodPreset
metadata:
  name: pki-volumes
spec:
  selector:
    matchLabels:
      inject-pki: "yes"
  volumeMounts:
    - mountPath: /etc/pki/ca-trust/extracted/pem
      name: root-ca-bundle-pems
      readOnly: true
    - mountPath: /etc/pki/ca-trust/extracted/java
      name: root-jks
      readOnly: true
  volumes:
    - configMap:
        items:
          - key: ca-bundle.crt
            path: tls-ca-bundle.pem
        name: root-ca-bundle-pems
      name: root-ca-bundle-pems
    - configMap:
        items:
          - key: cacerts
            path: cacerts
        name: root-jks
      name: root-jks
```

And a test Deployment would be something like:

```yaml
kind: Deployment
apiVersion: apps/v1
metadata:
  name: pki-toolbox
spec:
  replicas: 1
  selector:
    matchLabels:
      app: pki
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: pki
        inject-pki: "yes"
    spec:
      containers:
        - name: pki
          image: 'quay.io/kenmoini/pki-toolbox:latest'
          command:
            - /bin/bash
            - '-c'
            - '--'
          args:
            - while true; do sleep 30; done;
```

Extra information around additional options for PKI in OpenShift here: https://kenmoini.com/post/2022/02/custom-root-ca-in-openshift/

---

# OpenShift MutatingWebHook, PKI Mounter

This repository provides the needed constructs for a MutatingWebHook that will automatically attach ConfigMaps to Pods.  It can be used to automatically inject Root CA Bundle Certificate ConfigMaps into Pods.

## Deploy

1. Create a namespace `pki-injector` in which the PKI injector webhook is deployed:

```bash
oc new-project pki-injector
```

2. Deploy resources:

```bash
oc apply -n pki-injector -f openshift/
```

## Test & Verify

1. The sidecar inject webhook should be in running state:

```
oc -n pki-injector get pod
NAME                                                   READY   STATUS    RESTARTS   AGE
pki-injector-webhook-deployment-7c8bc5f4c9-28c84   1/1     Running   0          30s

oc -n pki-injector get deploy
NAME                                  READY   UP-TO-DATE   AVAILABLE   AGE
pki-injector-webhook-deployment   1/1     1            1           67s
```

2. Create new namespace `injection` and label it with `pki-injector=enabled`:

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
pki-injector     Active   17m
```

3. Deploy an app in the OpenShift cluster, take `busybox` app as an example

```
oc run busybox --image=busybox --restart=Never -n injection --overrides='{"apiVersion":"v1","metadata":{"annotations":{"pki-injector-webhook.polyglot.systems/inject":"yes"}}}' --command -- sleep infinity
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
oc delete namespace pki-injector
oc delete namespace injection

oc delete MutatingWebhookConfiguration/pki-injector.polyglot.systems
oc delete ValidatingWebhookConfiguration/pki-injector.polyglot.systems
```

## Local Development

```bash
## Log into the Container Registry
sudo podman login -u=admin -p=adminPass quay.io

## Build
sudo podman build -t quay.io/kenmoini/ocp-mutating-webhook-pod-mounter:latest .

## Push
sudo podman push quay.io/kenmoini/ocp-mutating-webhook-pod-mounter:latest
```
