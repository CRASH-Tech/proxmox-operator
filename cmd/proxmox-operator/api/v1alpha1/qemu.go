package v1alpha1

type QemuImpl struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Spec       struct {
		Accepted bool   `json:"accepted"`
		Cluster  string `json:"cluster"`
		Config   struct {
			Agent   bool   `json:"agent"`
			Cores   int    `json:"cores"`
			Sockets int    `json:"sockets"`
			Test    string `json:"test"`
		} `json:"config"`
		Node string `json:"node"`
		Pool string `json:"pool"`
	} `json:"spec"`
}
