package main

import (
	"context"
	"encoding/json"
	"fmt"

	proxmoxoperator "github.com/CRASH-Tech/proxmox-operator/cmd/proxmox-operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

// if kubeconfig == "" {
// 	log.Printf("using in-cluster configuration")
// 	config, err = rest.InClusterConfig()
// } else {
// 	log.Printf("using configuration from '%s'", kubeconfig)
// 	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
// }

func main() {
	// ctx := context.Background()
	// config := ctrl.GetConfigOrDie()
	// dynamic := dynamic.NewForConfigOrDie(config)

	// namespace := "sidero-system"

	// items, err := GetResourcesDynamically(dynamic, ctx, "proxmox.xfix.org", "v1alpha1", "qemu", namespace)
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	for _, item := range items {
	// 		fmt.Printf("%+v\n", item)
	// 	}
	// }

	// err := GetResourceDynamically(dynamic, ctx, "proxmox.xfix.org", "v1alpha1", "qemu", namespace)
	// fmt.Println(err)

	// err := PatchResourcesDynamically(dynamic, ctx, "proxmox.xfix.org", "v1alpha1", "qemu", namespace)
	// fmt.Println(err)

	proxmoxoperator.Loop()
}

func GetResourcesDynamically(dynamic dynamic.Interface, ctx context.Context,
	group string, version string, resource string, namespace string) (
	[]unstructured.Unstructured, error) {

	resourceId := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}
	list, err := dynamic.Resource(resourceId).Namespace(namespace).
		List(ctx, metav1.ListOptions{})

	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

func GetResourceDynamically(dynamic dynamic.Interface, ctx context.Context,
	group string, version string, resource string, namespace string) error {

	resourceId := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}

	//obj, err := dynamic.Resource(resourceId).Get(ctx, "example-qemu", metav1.GetOptions{})
	err := dynamic.Resource(resourceId).Delete(ctx, "example-qemu", metav1.DeleteOptions{})

	return err
}

func PatchResourcesDynamically(dynamic dynamic.Interface, ctx context.Context,
	group string, version string, resource string, namespace string) error {

	resourceId := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}

	patch := []interface{}{
		map[string]interface{}{
			"op":    "replace",
			"path":  "/spec/accepted",
			"value": true,
		},
	}

	payload, err := json.Marshal(patch)
	if err != nil {
		return err
	}

	list, err := dynamic.Resource(resourceId).Patch(ctx, "example-qemu", types.JSONPatchType, payload, metav1.PatchOptions{})

	fmt.Println(list)

	if err != nil {
		return err
	}

	return nil
}
