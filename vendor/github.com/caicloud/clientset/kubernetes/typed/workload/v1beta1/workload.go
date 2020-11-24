/*
Copyright 2019 caicloud authors. All rights reserved.
*/

// Code generated by client-gen. DO NOT EDIT.

package v1beta1

import (
	scheme "github.com/caicloud/clientset/kubernetes/scheme"
	v1beta1 "github.com/caicloud/clientset/pkg/apis/workload/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// WorkloadsGetter has a method to return a WorkloadInterface.
// A group's client should implement this interface.
type WorkloadsGetter interface {
	Workloads(namespace string) WorkloadInterface
}

// WorkloadInterface has methods to work with Workload resources.
type WorkloadInterface interface {
	Create(*v1beta1.Workload) (*v1beta1.Workload, error)
	Update(*v1beta1.Workload) (*v1beta1.Workload, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1beta1.Workload, error)
	List(opts v1.ListOptions) (*v1beta1.WorkloadList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.Workload, err error)
	WorkloadExpansion
}

// workloads implements WorkloadInterface
type workloads struct {
	client rest.Interface
	ns     string
}

// newWorkloads returns a Workloads
func newWorkloads(c *WorkloadV1beta1Client, namespace string) *workloads {
	return &workloads{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the workload, and returns the corresponding workload object, and an error if there is any.
func (c *workloads) Get(name string, options v1.GetOptions) (result *v1beta1.Workload, err error) {
	result = &v1beta1.Workload{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("workloads").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Workloads that match those selectors.
func (c *workloads) List(opts v1.ListOptions) (result *v1beta1.WorkloadList, err error) {
	result = &v1beta1.WorkloadList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("workloads").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested workloads.
func (c *workloads) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("workloads").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a workload and creates it.  Returns the server's representation of the workload, and an error, if there is any.
func (c *workloads) Create(workload *v1beta1.Workload) (result *v1beta1.Workload, err error) {
	result = &v1beta1.Workload{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("workloads").
		Body(workload).
		Do().
		Into(result)
	return
}

// Update takes the representation of a workload and updates it. Returns the server's representation of the workload, and an error, if there is any.
func (c *workloads) Update(workload *v1beta1.Workload) (result *v1beta1.Workload, err error) {
	result = &v1beta1.Workload{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("workloads").
		Name(workload.Name).
		Body(workload).
		Do().
		Into(result)
	return
}

// Delete takes name of the workload and deletes it. Returns an error if one occurs.
func (c *workloads) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("workloads").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *workloads) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("workloads").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched workload.
func (c *workloads) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.Workload, err error) {
	result = &v1beta1.Workload{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("workloads").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
