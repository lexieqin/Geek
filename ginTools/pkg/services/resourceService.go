package services

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/scheme"
)

// ResourceService provides methods to interact with Kubernetes resources
type ResourceService struct {
	restMapper *meta.RESTMapper
	client     *dynamic.DynamicClient
	fact       informers.SharedInformerFactory
}

// NewResourceService creates a new ResourceService
func NewResourceService(restMapper *meta.RESTMapper, client *dynamic.DynamicClient, fact informers.SharedInformerFactory) *ResourceService {
	return &ResourceService{restMapper: restMapper, client: client, fact: fact}
}

// ResourceInfo contains information about a Kubernetes resource
type ResourceInfo struct {
	Name        string                 `json:"name"`
	Namespace   string                 `json:"namespace"`
	Kind        string                 `json:"kind"`
	APIVersion  string                 `json:"apiVersion"`
	UID         string                 `json:"uid"`
	CreatedAt   metav1.Time            `json:"createdAt"`
	Labels      map[string]string      `json:"labels,omitempty"`
	Annotations map[string]string      `json:"annotations,omitempty"`
	Spec        map[string]interface{} `json:"spec,omitempty"`
	Status      map[string]interface{} `json:"status,omitempty"`
}

// ListResource lists resources of the specified type in the given namespace
func (r *ResourceService) ListResource(resourceOrKindArg string, ns string) ([]ResourceInfo, error) {
	restMapping, err := r.mappingFor(resourceOrKindArg, r.restMapper)
	if err != nil {
		return nil, fmt.Errorf("failed to map resource '%s': %w", resourceOrKindArg, err)
	}

	informer, err := r.fact.ForResource(restMapping.Resource)
	if err != nil {
		return nil, fmt.Errorf("failed to get informer for resource '%s': %w", restMapping.Resource.String(), err)
	}

	list, err := informer.Lister().ByNamespace(ns).List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	result := make([]ResourceInfo, 0, len(list))
	for _, obj := range list {
		unstructObj, ok := obj.(*unstructured.Unstructured)
		if !ok {
			continue
		}

		resourceInfo := ResourceInfo{
			Name:        unstructObj.GetName(),
			Namespace:   unstructObj.GetNamespace(),
			Kind:        unstructObj.GetKind(),
			APIVersion:  unstructObj.GetAPIVersion(),
			UID:         string(unstructObj.GetUID()),
			CreatedAt:   unstructObj.GetCreationTimestamp(),
			Labels:      unstructObj.GetLabels(),
			Annotations: unstructObj.GetAnnotations(),
		}

		// Extract spec and status if available
		if spec, found, _ := unstructured.NestedMap(unstructObj.Object, "spec"); found {
			resourceInfo.Spec = spec
		}
		if status, found, _ := unstructured.NestedMap(unstructObj.Object, "status"); found {
			resourceInfo.Status = status
		}

		result = append(result, resourceInfo)
	}

	return result, nil
}

// DeleteResource deletes a resource by name
func (r *ResourceService) DeleteResource(resourceOrKindArg string, ns string, name string) error {
	if name == "" {
		return fmt.Errorf("resource name cannot be empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ri, err := r.getResourceInterface(resourceOrKindArg, ns, r.client, r.restMapper)
	if err != nil {
		return fmt.Errorf("failed to get resource interface: %w", err)
	}

	err = ri.Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete resource '%s': %w", name, err)
	}
	return nil
}

// CreateResource creates a resource from YAML
func (r *ResourceService) CreateResource(resourceOrKindArg string, yaml string) error {
	if yaml == "" {
		return fmt.Errorf("YAML content cannot be empty")
	}

	obj := &unstructured.Unstructured{}
	_, _, err := scheme.Codecs.UniversalDeserializer().Decode([]byte(yaml), nil, obj)
	if err != nil {
		return fmt.Errorf("failed to decode YAML: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ri, err := r.getResourceInterface(resourceOrKindArg, obj.GetNamespace(), r.client, r.restMapper)
	if err != nil {
		return fmt.Errorf("failed to get resource interface: %w", err)
	}

	_, err = ri.Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create resource '%s': %w", obj.GetName(), err)
	}
	return nil
}

// GVRInfo contains information about a GroupVersionResource
type GVRInfo struct {
	Group      string `json:"group"`
	Version    string `json:"version"`
	Resource   string `json:"resource"`
	Kind       string `json:"kind,omitempty"`
	Namespaced bool   `json:"namespaced"`
}

// GetGVR returns the GroupVersionResource for a resource type
func (r *ResourceService) GetGVR(resourceOrKindArg string) (*GVRInfo, error) {
	if resourceOrKindArg == "" {
		return nil, fmt.Errorf("resource argument cannot be empty")
	}

	restMapping, err := r.mappingFor(resourceOrKindArg, r.restMapper)
	if err != nil {
		return nil, fmt.Errorf("failed to get mapping for '%s': %w", resourceOrKindArg, err)
	}

	return &GVRInfo{
		Group:      restMapping.Resource.Group,
		Version:    restMapping.Resource.Version,
		Resource:   restMapping.Resource.Resource,
		Kind:       restMapping.GroupVersionKind.Kind,
		Namespaced: restMapping.Scope.Name() == "namespace",
	}, nil
}

// getResourceInterface returns the appropriate ResourceInterface for the resource type
func (r *ResourceService) getResourceInterface(resourceOrKindArg string, ns string, client dynamic.Interface, restMapper *meta.RESTMapper) (dynamic.ResourceInterface, error) {
	restMapping, err := r.mappingFor(resourceOrKindArg, restMapper)
	if err != nil {
		return nil, fmt.Errorf("failed to get RESTMapping for %s: %w", resourceOrKindArg, err)
	}

	// Check if resource is namespaced or cluster-scoped
	if restMapping.Scope.Name() == "namespace" {
		return client.Resource(restMapping.Resource).Namespace(ns), nil
	}
	return client.Resource(restMapping.Resource), nil
}

// mappingFor gets the REST mapping for a resource or kind
func (r *ResourceService) mappingFor(resourceOrKindArg string, restMapper *meta.RESTMapper) (*meta.RESTMapping, error) {
	fullySpecifiedGVR, groupResource := schema.ParseResourceArg(resourceOrKindArg)
	gvk := schema.GroupVersionKind{}

	if fullySpecifiedGVR != nil {
		gvk, _ = (*restMapper).KindFor(*fullySpecifiedGVR)
	}
	if gvk.Empty() {
		fmt.Println("groupResource: ", groupResource)
		gvk, _ = (*restMapper).KindFor(groupResource.WithVersion(""))
		fmt.Println("gvk: ", gvk)
	}
	if !gvk.Empty() {
		return (*restMapper).RESTMapping(gvk.GroupKind(), gvk.Version)
	}

	fullySpecifiedGVK, groupKind := schema.ParseKindArg(resourceOrKindArg)
	if fullySpecifiedGVK == nil {
		gvk := groupKind.WithVersion("")
		fullySpecifiedGVK = &gvk
	}

	if !fullySpecifiedGVK.Empty() {
		if mapping, err := (*restMapper).RESTMapping(fullySpecifiedGVK.GroupKind(), fullySpecifiedGVK.Version); err == nil {
			return mapping, nil
		}
	}

	mapping, err := (*restMapper).RESTMapping(groupKind, gvk.Version)
	if err != nil {
		if meta.IsNoMatchError(err) {
			return nil, fmt.Errorf("the server doesn't have a resource type %q", groupResource.Resource)
		}
		return nil, err
	}

	return mapping, nil
}

// ResourceList contains information about a list of resources and their type
type ResourceList struct {
	Resources []ResourceInfo `json:"resources"`
	GVR       GVRInfo        `json:"gvr"`
}

// GetResource returns all resources of the specified type across all namespaces
func (r *ResourceService) GetResource(resource string) (*ResourceList, error) {
	if resource == "" {
		return nil, fmt.Errorf("resource argument cannot be empty")
	}

	restMapping, err := r.mappingFor(resource, r.restMapper)
	if err != nil {
		return nil, fmt.Errorf("failed to get mapping for '%s': %w", resource, err)
	}

	informer, err := r.fact.ForResource(restMapping.Resource)
	if err != nil {
		return nil, fmt.Errorf("failed to get informer for resource '%s': %w", restMapping.Resource.String(), err)
	}

	list, err := informer.Lister().List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	resources := make([]ResourceInfo, 0, len(list))
	for _, obj := range list {
		unstructObj, ok := obj.(*unstructured.Unstructured)
		if !ok {
			continue
		}

		resourceInfo := ResourceInfo{
			Name:        unstructObj.GetName(),
			Namespace:   unstructObj.GetNamespace(),
			Kind:        unstructObj.GetKind(),
			APIVersion:  unstructObj.GetAPIVersion(),
			UID:         string(unstructObj.GetUID()),
			CreatedAt:   unstructObj.GetCreationTimestamp(),
			Labels:      unstructObj.GetLabels(),
			Annotations: unstructObj.GetAnnotations(),
		}

		resources = append(resources, resourceInfo)
	}

	return &ResourceList{
		Resources: resources,
		GVR: GVRInfo{
			Group:      restMapping.Resource.Group,
			Version:    restMapping.Resource.Version,
			Resource:   restMapping.Resource.Resource,
			Kind:       restMapping.GroupVersionKind.Kind,
			Namespaced: restMapping.Scope.Name() == "namespace",
		},
	}, nil
}

// GetResourceByType returns resources filtered by a specific type
func (r *ResourceService) GetResourceByType(resource string, resourceType string) (*ResourceList, error) {
	if resource == "" {
		return nil, fmt.Errorf("resource argument cannot be empty")
	}

	// If resourceType is empty, delegate to GetResource
	if resourceType == "" {
		return r.GetResource(resource)
	}

	restMapping, err := r.mappingFor(resource, r.restMapper)
	if err != nil {
		return nil, fmt.Errorf("failed to get mapping for '%s': %w", resource, err)
	}

	informer, err := r.fact.ForResource(restMapping.Resource)
	if err != nil {
		return nil, fmt.Errorf("failed to get informer for resource '%s': %w", restMapping.Resource.String(), err)
	}

	// Create a label selector for the resource type if provided
	var selector labels.Selector
	if resourceType != "" {
		selector = labels.SelectorFromSet(labels.Set{"type": resourceType})
	} else {
		selector = labels.Everything()
	}

	list, err := informer.Lister().List(selector)
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}

	resources := make([]ResourceInfo, 0, len(list))
	for _, obj := range list {
		unstructObj, ok := obj.(*unstructured.Unstructured)
		if !ok {
			continue
		}

		resourceInfo := ResourceInfo{
			Name:        unstructObj.GetName(),
			Namespace:   unstructObj.GetNamespace(),
			Kind:        unstructObj.GetKind(),
			APIVersion:  unstructObj.GetAPIVersion(),
			UID:         string(unstructObj.GetUID()),
			CreatedAt:   unstructObj.GetCreationTimestamp(),
			Labels:      unstructObj.GetLabels(),
			Annotations: unstructObj.GetAnnotations(),
		}

		resources = append(resources, resourceInfo)
	}

	return &ResourceList{
		Resources: resources,
		GVR: GVRInfo{
			Group:      restMapping.Resource.Group,
			Version:    restMapping.Resource.Version,
			Resource:   restMapping.Resource.Resource,
			Kind:       restMapping.GroupVersionKind.Kind,
			Namespaced: restMapping.Scope.Name() == "namespace",
		},
	}, nil
}

// UpdateResource updates an existing resource using the provided YAML
func (r *ResourceService) UpdateResource(resourceOrKindArg string, ns string, name string, yaml string) error {
	if yaml == "" {
		return fmt.Errorf("YAML content cannot be empty")
	}
	obj := &unstructured.Unstructured{}
	_, _, err := scheme.Codecs.UniversalDeserializer().Decode([]byte(yaml), nil, obj)
	if err != nil {
		return fmt.Errorf("failed to decode YAML: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ri, err := r.getResourceInterface(resourceOrKindArg, ns, r.client, r.restMapper)
	if err != nil {
		return fmt.Errorf("failed to get resource interface: %w", err)
	}
	_, err = ri.Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update resource '%s': %w", name, err)
	}
	return nil
}

// PatchResource patches a resource using the provided patch string
func (r *ResourceService) PatchResource(resourceOrKindArg string, ns string, name string, patch string) error {
	if patch == "" {
		return fmt.Errorf("patch content cannot be empty")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ri, err := r.getResourceInterface(resourceOrKindArg, ns, r.client, r.restMapper)
	if err != nil {
		return fmt.Errorf("failed to get resource interface: %w", err)
	}
	_, err = ri.Patch(ctx, name, types.JSONPatchType, []byte(patch), metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("failed to patch resource '%s': %w", name, err)
	}
	return nil
}

// GetResourceStatus returns the status of a resource
func (r *ResourceService) GetResourceStatus(resourceOrKindArg string, ns string, name string) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ri, err := r.getResourceInterface(resourceOrKindArg, ns, r.client, r.restMapper)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource interface: %w", err)
	}
	obj, err := ri.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get resource '%s': %w", name, err)
	}
	status, found, err := unstructured.NestedMap(obj.Object, "status")
	if err != nil || !found {
		return nil, fmt.Errorf("failed to extract status: %w", err)
	}
	return status, nil
}
