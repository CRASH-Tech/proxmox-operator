# proxmox-operator

This repository provides a Kubernetes operator for managing [Proxmox virtualization platform](https://pve.proxmox.com/pve-docs/) QEMU VMs via CRDs.

## Getting Started

* `git clone https://github.com/CRASH-Tech/proxmox-operator.git`
* `cd proxmox-operator`
* Edit charts/proxmox-operator/values.yaml for your environment
* `helm repo add crash-tech https://crash-tech.github.io/charts/`
* `helm install proxmox-operator crash-tech/proxmox-operator -f charts/proxmox-operator/values.yaml`

## Deploy example VM

* Edit examples/example-qemu.yaml for your environment
* `kubectl apply -f examples/example-qemu.yaml`
* Check VM status
* `kubectl get qemu`

## Example VM
```
apiVersion: proxmox.xfix.org/v1alpha1
kind: Qemu
metadata:
  name: example-qemu
  finalizers:
    - resources-finalizer.proxmox-operator.xfix.org
spec:
  cluster: pve-test
  #node: crash-lab ### If not set it will set automaticly from "pool"
  #vmid: 222  ### If not set it will set automaticly
  pool: prod ### Cluster pool for place VM
  anti-affinity: "" ### The anti-affinity group. VM's with same anti-affinity group will be placed on different nodes
  autostart: true
  autostop: true
  cpu:
    type: host
    sockets: 2
    cores: 1
  memory:
    size: 2048
    balloon: 2048
  network:
    net0:
      model: virtio
      #mac: A2:7B:45:48:9C:E6  ### If not set it will set automaticly
      bridge: vmbr0
      tag: 103
  disk:
    scsi0:
      storage: local-lvm
      size: 9G
  tags:
    - test1
    - test2
  options:
    ostype: "l26"
    bios: "seabios"
    smbios1: "uuid=3ae878b3-a77e-4a4a-adc6-14ee88350d36,manufacturer=MTIz,product=MTIz,version=MTIz,serial=MTIz,sku=MTIz,family=MTIz,base64=1"
    scsihw: "virtio-scsi-pci"
    boot: "order=net0;ide2;scsi0"
    ide2: "none,media=cdrom"
    hotplug: "network,disk,usb"
    tablet: 1
    onboot: 0
    kvm: 1
    agent: "0"
    numa: 1
    protection: 0
```

## Useful links

* [Kubernetes CRD](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)
* [Proxmox](https://www.proxmox.com/en/)
* [Proxmox documentation](https://pve.proxmox.com/pve-docs/)
