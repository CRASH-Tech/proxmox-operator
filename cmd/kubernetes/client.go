package kuberentes

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

type Client struct {
	ctx     context.Context
	dynamic dynamic.DynamicClient
}

type V1alpha1 struct {
	client *Client
}

func NewClient(ctx context.Context, dynamic dynamic.DynamicClient) *Client {
	client := Client{
		ctx:     ctx,
		dynamic: dynamic,
	}

	return &client
}

func (client *Client) dynamicGet(resourceId schema.GroupVersionResource, name string) ([]byte, error) {

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

func (client *Client) dynamicGetAll(resourceId schema.GroupVersionResource) ([][]byte, error) {

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

func (client *Client) dynamicPatch(resourceId schema.GroupVersionResource, name string, patch []byte) ([]byte, error) {

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

func (client *Client) dynamicUpdateStatus(resourceId schema.GroupVersionResource, name string, patch []byte) ([]byte, error) {
	var data unstructured.Unstructured
	err := data.UnmarshalJSON(patch)
	if err != nil {
		return nil, err
	}

	result, err := client.dynamic.Resource(resourceId).UpdateStatus(client.ctx, &data, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	jsonData, err := result.MarshalJSON()
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

func (v1alpha1 *V1alpha1) Qemu() *Qemu {
	qemu := Qemu{
		client: v1alpha1.client,
		resourceId: schema.GroupVersionResource{
			Group:    "proxmox.xfix.org",
			Version:  "v1alpha1",
			Resource: "qemu",
		},
	}

	return &qemu
}
