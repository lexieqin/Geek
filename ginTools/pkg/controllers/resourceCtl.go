package controllers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/lexieqin/Geek/ginTools/pkg/services"
)

type ResourceCtl struct {
	resourceService *services.ResourceService
}

func NewResourceCtl(service *services.ResourceService) *ResourceCtl {
	return &ResourceCtl{resourceService: service}
}

func (r *ResourceCtl) List() func(c *gin.Context) {
	return func(c *gin.Context) {
		var resource = c.Param("resource")
		ns := c.DefaultQuery("ns", "default")
		resourceList, _ := r.resourceService.ListResource(resource, ns)
		c.JSON(200, gin.H{"data": resourceList})
	}
}

func (r *ResourceCtl) Delete() func(c *gin.Context) {
	return func(c *gin.Context) {
		var resource = c.Param("resource")
		ns := c.DefaultQuery("ns", "default")
		name := c.Query("name")
		err := r.resourceService.DeleteResource(resource, ns, name)
		if err != nil {
			c.JSON(500, gin.H{"error": "Delete failed: " + err.Error()})
			return
		} else {
			c.JSON(200, gin.H{"data": "Delete successful"})
		}
	}
}

func (r *ResourceCtl) Create() func(c *gin.Context) {
	fmt.Println("create")
	return func(c *gin.Context) {
		var resource = c.Param("resource")

		type ResouceParam struct {
			Yaml string `json:"yaml"`
		}

		var param ResouceParam
		if err := c.ShouldBindJSON(&param); err != nil {
			c.JSON(400, gin.H{"error": "Failed to parse request body: " + err.Error()})
			return
		}

		err := r.resourceService.CreateResource(resource, param.Yaml)
		if err != nil {
			c.JSON(400, gin.H{"error": "Creation failed: " + err.Error()})
			return
		} else {
			c.JSON(200, gin.H{"data": "Creation successful"})
		}
	}
}

func (r *ResourceCtl) GetGVR() func(c *gin.Context) {
	return func(c *gin.Context) {
		var resource = c.Query("resource")

		gvr, err := r.resourceService.GetGVR(resource)
		if err != nil {
			c.JSON(400, gin.H{"error": "Resource error: " + err.Error()})
			return
		} else {
			c.JSON(200, gin.H{"data": *gvr})
		}
	}
}

func (r *ResourceCtl) GetResource() gin.HandlerFunc {
	return func(c *gin.Context) {
		var resource = c.Query("resource")

		resourceList, err := r.resourceService.GetResource(resource)
		if err != nil {
			c.JSON(400, gin.H{"error": "Resource error: " + err.Error()})
			return
		}
		c.JSON(200, gin.H{"data": resourceList})
	}
}

func (r *ResourceCtl) GetResourceByType() gin.HandlerFunc {
	return func(c *gin.Context) {
		var resource = c.Query("resource")
		var resourceType = c.Query("type")

		resourceList, err := r.resourceService.GetResourceByType(resource, resourceType)
		if err != nil {
			c.JSON(400, gin.H{"error": "Resource error: " + err.Error()})
			return
		}
		c.JSON(200, gin.H{"data": resourceList})
	}
}

// Update updates an existing resource
func (r *ResourceCtl) Update() gin.HandlerFunc {
	return func(c *gin.Context) {
		var resource = c.Param("resource")
		ns := c.DefaultQuery("ns", "default")
		name := c.Query("name")
		if name == "" {
			c.JSON(400, gin.H{"error": "name parameter is required"})
			return
		}
		var yaml string
		if err := c.ShouldBindJSON(&gin.H{"yaml": &yaml}); err != nil {
			c.JSON(400, gin.H{"error": "Failed to parse request body: " + err.Error()})
			return
		}
		err := r.resourceService.UpdateResource(resource, ns, name, yaml)
		if err != nil {
			c.JSON(500, gin.H{"error": "Update failed: " + err.Error()})
			return
		}
		c.JSON(200, gin.H{"data": "Update successful"})
	}
}

// Patch patches a resource
func (r *ResourceCtl) Patch() gin.HandlerFunc {
	return func(c *gin.Context) {
		var resource = c.Param("resource")
		ns := c.DefaultQuery("ns", "default")
		name := c.Query("name")
		if name == "" {
			c.JSON(400, gin.H{"error": "name parameter is required"})
			return
		}
		var patch string
		if err := c.ShouldBindJSON(&gin.H{"patch": &patch}); err != nil {
			c.JSON(400, gin.H{"error": "Failed to parse request body: " + err.Error()})
			return
		}
		err := r.resourceService.PatchResource(resource, ns, name, patch)
		if err != nil {
			c.JSON(500, gin.H{"error": "Patch failed: " + err.Error()})
			return
		}
		c.JSON(200, gin.H{"data": "Patch successful"})
	}
}

// GetStatus returns the status of a resource
func (r *ResourceCtl) GetStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		var resource = c.Param("resource")
		ns := c.DefaultQuery("ns", "default")
		name := c.Query("name")
		if name == "" {
			c.JSON(400, gin.H{"error": "name parameter is required"})
			return
		}
		status, err := r.resourceService.GetResourceStatus(resource, ns, name)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to get status: " + err.Error()})
			return
		}
		c.JSON(200, gin.H{"data": status})
	}
}
