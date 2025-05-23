package controllers

// PodLogEventCtl 用于处理 Pod 日志和事件的控制器
//

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lexieqin/Geek/ginTools/pkg/services"
)

type PodLogEventCtl struct {
	podLogEventService *services.PodLogEventService
}

func NewPodLogEventCtl(service *services.PodLogEventService) *PodLogEventCtl {
	return &PodLogEventCtl{podLogEventService: service}
}

func (p *PodLogEventCtl) GetLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		ns := c.Param("namespace")
		podName := c.Param("podName")
		containerName := c.DefaultQuery("container", "")
		tailLineStr := c.DefaultQuery("tail", "100")

		// Convert tailLine from string to int64
		tailLine, err := strconv.ParseInt(tailLineStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Invalid tail parameter: %v", err),
			})
			return
		}

		// Validate required parameters
		if podName == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "podname parameter is required",
			})
			return
		}

		// Create a context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Get logs with the improved service
		req := p.podLogEventService.GetLogs(ns, podName, tailLine, containerName)

		// Stream the logs
		rc, err := req.Stream(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to stream logs: %v", err),
			})
			return
		}
		defer rc.Close()

		// Read log data with a buffer size limit to prevent memory issues
		logData, err := io.ReadAll(rc)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to read logs: %v", err),
			})
			return
		}

		// If the client accepts text/plain, return as plain text
		if c.NegotiateFormat(gin.MIMEJSON, gin.MIMEPlain) == gin.MIMEPlain {
			c.String(http.StatusOK, string(logData))
			return
		}

		// Otherwise return as JSON
		c.JSON(http.StatusOK, gin.H{
			"data": string(logData),
			"meta": gin.H{
				"namespace":     ns,
				"podName":       podName,
				"containerName": containerName,
				"tailLines":     tailLine,
			},
		})
	}
}

func (p *PodLogEventCtl) GetEvent() gin.HandlerFunc {
	return func(c *gin.Context) {
		ns := c.Param("namespace")
		podName := c.Param("podName")
		// Add support for filtering by event type (optional parameter)
		eventType := c.DefaultQuery("type", "")

		if podName == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "podname parameter is required",
			})
			return
		}

		events, err := p.podLogEventService.GetEvents(ns, podName, eventType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to retrieve events: %v", err),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": events,
			"meta": gin.H{
				"namespace": ns,
				"podName":   podName,
				"count":     len(events),
			},
		})
	}
}

func (p *PodLogEventCtl) ListPods() gin.HandlerFunc {
	return func(c *gin.Context) {
		ns := c.Param("namespace")
		podList, err := p.podLogEventService.ListPods(ns)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(200, gin.H{"data": podList})
	}
}

func (p *PodLogEventCtl) GetPod() gin.HandlerFunc {
	return func(c *gin.Context) {
		ns := c.Param("namespace")
		podName := c.Param("podName")

		pod, err := p.podLogEventService.GetPod(ns, podName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(200, gin.H{"data": pod})
	}
}
