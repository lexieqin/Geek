package main

import (
	"github.com/gin-gonic/gin"
	"github.com/xingyunyang01/ginTools/pkg/config"
	"github.com/xingyunyang01/ginTools/pkg/controllers"
	"github.com/xingyunyang01/ginTools/pkg/services"
)

func main() {
	k8sconfig := config.NewK8sConfig().InitRestConfig()
	restMapper := k8sconfig.InitRestMapper()
	dynamicClient := k8sconfig.InitDynamicClient()
	informer := k8sconfig.InitInformer()

	clientSet := k8sconfig.InitClientSet()

	resourceCtl := controllers.NewResourceCtl(services.NewResourceService(&restMapper, dynamicClient, informer))
	podLogCtl := controllers.NewPodLogEventCtl(services.NewPodLogEventService(clientSet))

	r := gin.New()

	r.GET("/:resource", resourceCtl.List())
	r.DELETE("/:resource", resourceCtl.Delete())
	r.POST("/:resource", resourceCtl.Create())
	r.PUT("/:resource", resourceCtl.Update())
	r.PATCH("/:resource", resourceCtl.Patch())
	r.GET("/:resource/status", resourceCtl.GetStatus())
	r.GET("/get/gvr", resourceCtl.GetGVR())
	r.GET("/get/resource", resourceCtl.GetResource())
	r.POST("/get/resource", resourceCtl.GetResourceByType())

	// Handle pod logs and events
	r.GET("/namespaces/:namespace/pods", podLogCtl.ListPods())
	r.GET("/namespaces/:namespace/pods/:podName", podLogCtl.GetPod())
	r.GET("/namespaces/:namespace/pods/:podName/logs", podLogCtl.GetLog())
	r.GET("/namespaces/:namespace/pods/:podName/events", podLogCtl.GetEvent())

	// Handle all pod logs and events
	r.GET("/namespaces/:namespace/pods/logs", podLogCtl.GetLog())
	r.GET("/namespaces/:namespace/pods/events", podLogCtl.GetEvent())

	r.Run(":8080")
}
