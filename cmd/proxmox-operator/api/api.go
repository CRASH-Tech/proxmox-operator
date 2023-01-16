package api

import (
	"context"
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

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
