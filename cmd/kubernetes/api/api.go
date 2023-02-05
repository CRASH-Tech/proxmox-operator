package api

type CustomResource struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Name                       string   `json:"name"`
		Uid                        string   `json:"uid"`
		Generation                 int      `json:"generation"`
		ResourceVersion            string   `json:"resourceVersion"`
		CreationTimestamp          string   `json:"creationTimestamp"`
		DeletionGracePeriodSeconds int      `json:"deletionGracePeriodSeconds,omitempty"`
		DeletionTimestamp          string   `json:"deletionTimestamp,omitempty"`
		Finalizers                 []string `json:"finalizers"`
	} `json:"metadata"`
}

func (cr *CustomResource) RemoveFinalizers() {
	cr.Metadata.Finalizers = []string{}
}
