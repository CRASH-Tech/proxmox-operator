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

package api

type CustomResource struct {
	APIVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Metadata   CustomResourceMetadata `json:"metadata"`
}

type CustomResourceMetadata struct {
	Name                       string   `json:"name"`
	Uid                        string   `json:"uid"`
	Generation                 int      `json:"generation"`
	ResourceVersion            string   `json:"resourceVersion"`
	CreationTimestamp          string   `json:"creationTimestamp"`
	DeletionGracePeriodSeconds int      `json:"deletionGracePeriodSeconds,omitempty"`
	DeletionTimestamp          string   `json:"deletionTimestamp,omitempty"`
	Finalizers                 []string `json:"finalizers"`
}

func (cr *CustomResource) RemoveFinalizers() {
	cr.Metadata.Finalizers = []string{}
}
