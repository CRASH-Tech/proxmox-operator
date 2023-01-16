package api

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
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

func DynamicUpdateClusterResource(ctx context.Context, dynamic dynamic.Interface,
	resourceId schema.GroupVersionResource, name string, obj unstructured.Unstructured) (unstructured.Unstructured, error) {

	item, err := dynamic.Resource(resourceId).Update(ctx, &obj, metav1.UpdateOptions{})
	if err != nil {
		return unstructured.Unstructured{}, err
	}

	return *item, nil
}

func DynamicPatchClusterResource(ctx context.Context, dynamic dynamic.Interface,
	resourceId schema.GroupVersionResource, name string, options []byte) (unstructured.Unstructured, error) {

	//item, err := dynamic.Resource(resourceId).Patch(ctx, "example-qemu", types.JSONPatchType, payload, metav1.PatchOptions{})
	item, err := dynamic.Resource(resourceId).Patch(ctx, name, types.MergePatchType, options, metav1.PatchOptions{})

	if err != nil {
		return unstructured.Unstructured{}, err
	}

	return *item, nil
}
