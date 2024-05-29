/*
Copyright 2024 The CRASH-Tech.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"github.com/CRASH-Tech/proxmox-operator/cmd/proxmox"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type Config struct {
	Log              LogConfig                           `yaml:"log"`
	Clusters         map[string]proxmox.ClusterApiConfig `yaml:"clusters"`
	DynamicClient    *dynamic.DynamicClient
	KubernetesClient *kubernetes.Clientset
	Listen           string `yaml:"listen"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}
