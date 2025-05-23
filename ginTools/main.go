package main

import (
	"os"
	
	"github.com/gin-gonic/gin"
	"github.com/lexieqin/Geek/ginTools/pkg/config"
	"github.com/lexieqin/Geek/ginTools/pkg/controllers"
	"github.com/lexieqin/Geek/ginTools/pkg/services"
)

func main() {
	var k8sconfig *config.K8sConfig
	
	// Check environment variable
	if os.Getenv("K8S_CONFIG_TYPE") == "in-cluster" {
		// Force in-cluster config only
		k8sconfig = config.NewK8sConfig().InitConfigInCluster()
	} else {
		// Use auto-detection (in-cluster first, then kubeconfig)
		k8sconfig = config.NewK8sConfig().InitRestConfig()
	}
	restMapper := k8sconfig.InitRestMapper()
	dynamicClient := k8sconfig.InitDynamicClient()
	informer := k8sconfig.InitInformer()

	clientSet := k8sconfig.InitClientSet()

	resourceCtl := controllers.NewResourceCtl(services.NewResourceService(&restMapper, dynamicClient, informer))
	podLogCtl := controllers.NewPodLogEventCtl(services.NewPodLogEventService(clientSet))
	jobDebugCtl := controllers.NewJobDebugController(services.NewJobDebugService(clientSet))
	mockJobCtl := controllers.NewMockJobController()

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

	// Job debug endpoints
	r.GET("/jobs/:namespace/:name/debug", jobDebugCtl.GetJobDebugInfo)
	r.GET("/jobs/:namespace/:name/traces", jobDebugCtl.GetJobTraces)
	r.GET("/jobs/:namespace/:name/errors", jobDebugCtl.GetJobErrors)
	r.GET("/jobs/:namespace/:name/sandbox", jobDebugCtl.GetJobSandboxLogs)
	r.GET("/jobs/:namespace/:name/pods", jobDebugCtl.GetJobPods)
	r.GET("/jobs/uuid/:uuid", jobDebugCtl.GetJobByUUID)
	r.GET("/sandbox/read", jobDebugCtl.ReadSandboxLog)

	// Mock endpoints for testing with real job data
	r.GET("/mock/jobs/:jobid/debug", mockJobCtl.GetMockJob)
	r.GET("/mock/jobs/uuid/:uuid", mockJobCtl.GetMockJobByUUID)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
