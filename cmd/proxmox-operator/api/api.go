package api

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func DynamicGetClusterResources(ctx context.Context, dynamic dynamic.Interface,
	resourceId schema.GroupVersionResource) ([]unstructured.Unstructured, error) {

	items, err := dynamic.Resource(resourceId).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return items.Items, nil
}

func DynamicGetClusterResource(ctx context.Context, dynamic dynamic.Interface,
	resourceId schema.GroupVersionResource, name string) (unstructured.Unstructured, error) {

	obj, err := dynamic.Resource(resourceId).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return unstructured.Unstructured{}, err
	}

	return *obj, nil
}

// func PatchResourcesDynamically(dynamic dynamic.Interface, ctx context.Context,
// 	group string, version string, resource string, namespace string) error {

// 	resourceId := schema.GroupVersionResource{
// 		Group:    group,
// 		Version:  version,
// 		Resource: resource,
// 	}

// 	patch := []interface{}{
// 		map[string]interface{}{
// 			"op":    "replace",
// 			"path":  "/spec/accepted",
// 			"value": true,
// 		},
// 	}

// 	payload, err := json.Marshal(patch)
// 	if err != nil {
// 		return err
// 	}

// 	list, err := dynamic.Resource(resourceId).Patch(ctx, "example-qemu", types.JSONPatchType, payload, metav1.PatchOptions{})

// 	fmt.Println(list)

// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
