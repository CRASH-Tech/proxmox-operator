package api

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

type Api struct {
	Ctx     context.Context
	Dynamic dynamic.DynamicClient
}

func New(ctx context.Context, dynamic dynamic.DynamicClient) *Api {
	api := Api{
		Ctx:     ctx,
		Dynamic: dynamic,
	}

	return &api
}

func (api Api) DynamicGetClusterResource(ctx context.Context, dynamic dynamic.Interface,
	resourceId schema.GroupVersionResource, name string) ([]byte, error) {

	item, err := dynamic.Resource(resourceId).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	jsonData, err := item.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

func (api Api) DynamicGetClusterResources(ctx context.Context, dynamic dynamic.Interface,
	resourceId schema.GroupVersionResource) ([][]byte, error) {

	items, err := dynamic.Resource(resourceId).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var result [][]byte
	for _, item := range items.Items {
		jsonData, err := item.MarshalJSON()
		if err != nil {
			return nil, err
		}
		result = append(result, jsonData)
	}

	return result, nil
}

// func DynamicUpdateClusterResource(ctx context.Context, dynamic dynamic.Interface,
// 	resourceId schema.GroupVersionResource, name string, obj unstructured.Unstructured) (unstructured.Unstructured, error) {

// 	item, err := dynamic.Resource(resourceId).Update(ctx, &obj, metav1.UpdateOptions{})
// 	if err != nil {
// 		return unstructured.Unstructured{}, err
// 	}

// 	return *item, nil
// }

func (api Api) DynamicPatchClusterResource(ctx context.Context, dynamic dynamic.Interface,
	resourceId schema.GroupVersionResource, name string, patch []byte) ([]byte, error) {

	item, err := dynamic.Resource(resourceId).Patch(ctx, name, types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return nil, err
	}

	jsonData, err := item.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}
