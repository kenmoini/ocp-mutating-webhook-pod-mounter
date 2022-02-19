package mutation

import (
	"github.com/kenmoini/ocp-mutating-webhook-pod-mounter/pkg/shared"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

// injectConfigMap is a container for the mutation injecting environment vars
type injectConfigMap struct {
	Logger logrus.FieldLogger
}

// injectConfigMap implements the podMutator interface
var _ podMutator = (*injectConfigMap)(nil)

// Name returns the struct name
func (se injectConfigMap) Name() string {
	return "inject_configmap"
}

// Mutate returns a new mutated pod according to set ConfigMap rules
func (se injectConfigMap) Mutate(pod *corev1.Pod) (*corev1.Pod, error) {

	se.Logger = se.Logger.WithField("mutation", se.Name())
	mpod := pod.DeepCopy()

	// Get the Configuration
	cfgFile := shared.CfgFile
	se.Logger = se.Logger.WithField("configuration", cfgFile)

	// build out env var slice
	envVars := []corev1.EnvVar{{
		Name:  "KUBE",
		Value: "true",
	}}

	// inject env vars into pod
	for _, envVar := range envVars {
		se.Logger.Debugf("pod configmap injected %s", envVar)
		injectConfigMapKey(mpod, envVar)
	}

	return mpod, nil
}

// injectConfigMapVar injects a var in both containers and init containers of a pod
func injectConfigMapKey(pod *corev1.Pod, envVar corev1.EnvVar) {
	for i, container := range pod.Spec.Containers {
		if !HasConfigMap(container, envVar) {
			pod.Spec.Containers[i].Env = append(container.Env, envVar)
		}
	}
	for i, container := range pod.Spec.InitContainers {
		if !HasConfigMap(container, envVar) {
			pod.Spec.InitContainers[i].Env = append(container.Env, envVar)
		}
	}
}

// HasConfigMap returns true if environment variable exists false otherwise
func HasConfigMap(container corev1.Container, checkEnvVar corev1.EnvVar) bool {
	for _, envVar := range container.Env {
		if envVar.Name == checkEnvVar.Name {
			return true
		}
	}
	return false
}
