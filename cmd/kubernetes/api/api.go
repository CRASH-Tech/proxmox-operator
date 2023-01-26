package api

type CustomResource struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Name              string   `json:"name"`
		ResourceVersion   string   `json:"resourceVersion"`
		DeletionTimestamp string   `json:"deletionTimestamp,omitempty"`
		Finalizers        []string `json:"finalizers,omitempty"`
	} `json:"metadata"`
}

func (cr *CustomResource) RemoveFinalizers() {
	cr.Metadata.Finalizers = nil
}
