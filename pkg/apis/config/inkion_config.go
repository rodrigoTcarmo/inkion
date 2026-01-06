package config

import (
	corev1 "k8s.io/api/core/v1"
)

type InkionConfig struct {
	ExpectedSender string
}

func NewInkionFromConfigMap(configMap *corev1.ConfigMap) (*InkionConfig, error) {
	inkionConfig := &InkionConfig{}
	for configName, configValue := range configMap.Data {
		if configName == "expected-sender" {
			inkionConfig.ExpectedSender = configValue
		}
	}

	return inkionConfig, nil
}
