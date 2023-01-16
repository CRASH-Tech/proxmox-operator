package v1alpha1

type QemuImpl struct {
	APIVersion string `json:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty"`
	Metadata   struct {
		Name string `json:"name,omitempty"`
	} `json:"metadata,omitempty"`
	Spec struct {
		Accepted bool   `json:"accepted,omitempty"`
		Cluster  string `json:"cluster,omitempty"`
		Config   struct {
			Agent   bool   `json:"agent,omitempty"`
			Cores   int    `json:"cores,omitempty"`
			Sockets int    `json:"sockets,omitempty"`
			Test    string `json:"test,omitempty"`
		} `json:"config,omitempty"`
		Node string `json:"node,omitempty"`
		Pool string `json:"pool,omitempty"`
	} `json:"spec,omitempty"`
}
