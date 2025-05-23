package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type MockJobController struct{}

func NewMockJobController() *MockJobController {
	return &MockJobController{}
}

// GetMockJob returns a real job response for testing
func (c *MockJobController) GetMockJob(ctx *gin.Context) {
	jobID := ctx.Param("jobid")
	
	// You can replace this with your real job data
	// This is just a template showing the structure
	mockResponse := map[string]interface{}{
		"job": map[string]interface{}{
			"name":      "etl-job-prod-" + jobID,
			"namespace": "production",
			"uuid":      jobID,
			"status":    "Failed",
			"labels": map[string]string{
				"job-uuid": jobID,
				"app":      "etl-processor",
				"env":      "production",
			},
		},
		"jobError": map[string]interface{}{
			"category":    "DATABASE_ERROR",
			"errorCode":   "ERR_CONN_REFUSED",
			"description": "Connection to database failed: dial tcp 10.0.1.5:5432: connect: connection refused",
			"timestamp":   "2024-01-20T15:30:45Z",
			"severity":    "CRITICAL",
		},
		"traces": map[string]interface{}{
			"datadogUrl": "https://app.datadoghq.com/apm/trace/8765432109876543210",
			"traceId":    "8765432109876543210",
			"spanId":     "1234567890123456",
			"traceLink":  "https://app.datadoghq.com/apm/trace/8765432109876543210?env=production",
		},
		"logs": map[string]interface{}{
			"sandboxPath": "/mnt/logs/jobs/production/etl-job-prod-" + jobID,
			"sandboxUrl":  "https://logs.internal.company.com/jobs/" + jobID,
			"logFiles": map[string]string{
				"stdout": "std.out",
				"stderr": "std.err",
				"decout": "decout",
				"decerr": "decerr",
			},
		},
		"errors": map[string]interface{}{
			"type":      "JobFailure",
			"reason":    "BackoffLimitExceeded",
			"message":   "Job has reached the specified backoff limit",
			"timestamp": "2024-01-20T15:31:00Z",
			"podErrors": []map[string]interface{}{
				{
					"podName":   "etl-job-prod-" + jobID + "-xj8kp",
					"container": "main",
					"reason":    "Error",
					"message":   "Container exited with status 1",
				},
			},
		},
		"events": []string{
			"[Normal] SuccessfulCreate: Created pod: etl-job-prod-" + jobID + "-xj8kp",
			"[Warning] BackoffLimitExceeded: Job has reached the specified backoff limit",
		},
		"pods": []map[string]interface{}{
			{
				"name":   "etl-job-prod-" + jobID + "-xj8kp",
				"status": "Failed",
				"node":   "worker-node-03",
			},
		},
	}

	ctx.JSON(http.StatusOK, mockResponse)
}

// GetMockJobByUUID returns a job by UUID for testing
func (c *MockJobController) GetMockJobByUUID(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	
	// Return the same mock data as if it was found by UUID
	mockResponse := map[string]interface{}{
		"apiVersion": "batch/v1",
		"kind":       "Job",
		"metadata": map[string]interface{}{
			"name":      "etl-job-prod-" + uuid,
			"namespace": "production",
			"labels": map[string]string{
				"job-uuid": uuid,
			},
			"annotations": map[string]string{
				"job.error.category":    "DATABASE_ERROR",
				"job.error.code":        "ERR_CONN_REFUSED",
				"job.error.description": "Connection to database failed: dial tcp 10.0.1.5:5432: connect: connection refused",
				"datadog.trace.id":      "8765432109876543210",
				"sandbox.path":          "/mnt/logs/jobs/production/etl-job-prod-" + uuid,
			},
		},
		"status": map[string]interface{}{
			"failed": 1,
			"conditions": []map[string]interface{}{
				{
					"type":               "Failed",
					"status":             "True",
					"reason":             "BackoffLimitExceeded",
					"message":            "Job has reached the specified backoff limit",
					"lastTransitionTime": "2024-01-20T15:31:00Z",
				},
			},
		},
	}
	
	ctx.JSON(http.StatusOK, mockResponse)
}