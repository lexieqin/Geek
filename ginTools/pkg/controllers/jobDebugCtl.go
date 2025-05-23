package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lexieqin/Geek/ginTools/pkg/services"
)

type JobDebugController struct {
	service *services.JobDebugService
}

func NewJobDebugController(service *services.JobDebugService) *JobDebugController {
	return &JobDebugController{service: service}
}

// GetJobDebugInfo returns comprehensive debug information for a job
func (c *JobDebugController) GetJobDebugInfo(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")

	debugInfo, err := c.service.GetJobDebugInfo(namespace, name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, debugInfo)
}

// GetJobByUUID finds a job by its UUID
func (c *JobDebugController) GetJobByUUID(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	namespace := ctx.Query("namespace") // optional namespace filter

	job, err := c.service.GetJobByUUID(uuid, namespace)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, job)
}

// GetJobTraces returns Datadog trace links for a job
func (c *JobDebugController) GetJobTraces(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")

	traces, err := c.service.GetJobTraces(namespace, name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, traces)
}

// GetJobErrors returns error details from the job
func (c *JobDebugController) GetJobErrors(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")

	errors, err := c.service.GetJobErrors(namespace, name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, errors)
}

// GetJobSandboxLogs returns sandbox logs for a job
func (c *JobDebugController) GetJobSandboxLogs(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")

	logs, err := c.service.GetJobSandboxLogs(namespace, name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, logs)
}

// GetJobPods returns all pods associated with a job
func (c *JobDebugController) GetJobPods(ctx *gin.Context) {
	namespace := ctx.Param("namespace")
	name := ctx.Param("name")

	pods, err := c.service.GetJobPods(namespace, name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, pods)
}

// ReadSandboxLog reads a specific log file from the sandbox directory
func (c *JobDebugController) ReadSandboxLog(ctx *gin.Context) {
	sandboxPath := ctx.Query("path")
	logFile := ctx.Query("file")
	startLine := ctx.DefaultQuery("start", "0")
	numLines := ctx.DefaultQuery("lines", "1000")

	if sandboxPath == "" || logFile == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "path and file parameters are required"})
		return
	}

	content, err := c.service.ReadSandboxLogFile(sandboxPath, logFile, startLine, numLines)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"path":    sandboxPath,
		"file":    logFile,
		"content": content,
	})
}
