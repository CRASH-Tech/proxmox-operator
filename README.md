# proxmox-operator

This repository provides kuberentes operator for manage the [Proxmox virtualization platform](https://pve.proxmox.com/pve-docs/) QEMU VMs via CRDs

## Getting Started


* `helm repo add crash-tech https://crash-tech.github.io/charts/`
* `helm install proxmox-operator crash-tech/proxmox-operator`

## Deploy example VM

* Edit examples/example-qemu.yaml for your environment
* `kubectl apply -f examples/example-qemu.yaml`

## Useful links

* [Kubernetes CRD](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
* [Proxmox](https://www.proxmox.com/en/)
* [Proxmox documentation](https://pve.proxmox.com/pve-docs/)
