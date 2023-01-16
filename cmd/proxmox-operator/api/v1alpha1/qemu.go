package v1alpha1

MOVE API HERE! AND CONVERT UNSTRUCTURED TO QEMU HERE!!!

type Qemu struct {
	APIVersion string    `json:"apiVersion"`
	Kind       string    `json:"kind"`
	Metadata   *Metadata `json:"metadata"`
	Spec       *Spec     `json:"spec"`
}
type Metadata struct {
	Name string `json:"name"`
}
type Config struct {
	Agent   bool   `json:"agent"`
	Cores   int    `json:"cores"`
	Sockets int    `json:"sockets"`
	Test    string `json:"test"`
}
type Spec struct {
	Accepted bool    `json:"accepted"`
	Cluster  string  `json:"cluster"`
	Config   *Config `json:"config"`
	Node     string  `json:"node"`
	Pool     string  `json:"pool"`
}
