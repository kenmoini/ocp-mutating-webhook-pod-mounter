package shared

import (
	"crypto/sha256"
	"io/ioutil"

	"github.com/golang/glog"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
)

var (
	CfgFile = &Config{}
)

type Config struct {
	Volumes      []corev1.Volume      `yaml:"volumes"`
	VolumeMounts []corev1.VolumeMount `yaml:"volumeMounts"`
}

func LoadConfig(configFile string) (*Config, error) {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	glog.Infof("New configuration: sha256sum %x", sha256.Sum256(data))

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
