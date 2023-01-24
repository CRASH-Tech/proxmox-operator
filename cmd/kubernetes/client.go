package kuberentes

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

type Client struct {
	ctx     context.Context
	dynamic dynamic.DynamicClient
}

func NewClient(ctx context.Context, dynamic dynamic.DynamicClient) *Client {
	client := Client{
		ctx:     ctx,
		dynamic: dynamic,
	}

	return &client
}

func (client *Client) dynamicGetClusterResource(resourceId schema.GroupVersionResource, name string) ([]byte, error) {

	item, err := client.dynamic.Resource(resourceId).Get(client.ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	jsonData, err := item.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

func (client *Client) dynamicGetClusterResources(resourceId schema.GroupVersionResource) ([][]byte, error) {

	items, err := client.dynamic.Resource(resourceId).List(client.ctx, metav1.ListOptions{})
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

func (client *Client) dynamicPatchClusterResource(resourceId schema.GroupVersionResource, name string, patch []byte) ([]byte, error) {

	item, err := client.dynamic.Resource(resourceId).Patch(client.ctx, name, types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return nil, err
	}

	jsonData, err := item.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

func (client *Client) V1alpha1() *V1alpha1 {
	result := V1alpha1{
		client: client,
	}

	return &result
}

// func (v1alpha1 *V1alpha1) Qemu() {
// 	fmt.Println(s)
// }
