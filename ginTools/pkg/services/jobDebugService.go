package services

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type JobDebugService struct {
	clientset kubernetes.Interface
}

func NewJobDebugService(clientset kubernetes.Interface) *JobDebugService {
	return &JobDebugService{clientset: clientset}
}

type JobDebugInfo struct {
	Job      *JobSummary  `json:"job"`
	JobError *JobError    `json:"jobError,omitempty"`
	Traces   *TraceInfo   `json:"traces"`
	Errors   *ErrorInfo   `json:"errors"`
	Logs     *LogInfo     `json:"logs"`
	Events   []string     `json:"events"`
	Pods     []PodSummary `json:"pods"`
}

type JobSummary struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	UUID      string            `json:"uuid"`
	Status    string            `json:"status"`
	Labels    map[string]string `json:"labels"`
}

type TraceInfo struct {
	DatadogURL string `json:"datadogUrl,omitempty"`
	TraceID    string `json:"traceId,omitempty"`
	SpanID     string `json:"spanId,omitempty"`
	TraceLink  string `json:"traceLink,omitempty"`
}

type ErrorInfo struct {
	Type      string     `json:"type"`
	Reason    string     `json:"reason"`
	Message   string     `json:"message"`
	Timestamp string     `json:"timestamp"`
	PodErrors []PodError `json:"podErrors,omitempty"`
}

type PodError struct {
	PodName   string `json:"podName"`
	Container string `json:"container"`
	Reason    string `json:"reason"`
	Message   string `json:"message"`
}

type LogInfo struct {
	SandboxPath string            `json:"sandboxPath,omitempty"`
	SandboxURL  string            `json:"sandboxUrl,omitempty"`
	LogFiles    map[string]string `json:"logFiles,omitempty"`
	Containers  map[string]string `json:"containers,omitempty"`
}

type PodSummary struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Node   string `json:"node"`
}

// GetJobDebugInfo returns comprehensive debug information for a job
func (s *JobDebugService) GetJobDebugInfo(namespace, name string) (*JobDebugInfo, error) {
	ctx := context.Background()

	// Get the job
	job, err := s.clientset.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	debugInfo := &JobDebugInfo{
		Job: s.getJobSummary(job),
	}

	// Extract JobError from annotations (pre-categorized by other services)
	debugInfo.JobError = s.extractJobError(job)

	// Get trace information
	debugInfo.Traces = s.extractTraceInfo(job)

	// Get error information
	debugInfo.Errors, err = s.getJobErrors(job)
	if err != nil {
		// Don't fail the whole request if we can't get errors
		debugInfo.Errors = &ErrorInfo{Message: fmt.Sprintf("Failed to get errors: %v", err)}
	}

	// Get associated pods
	pods, err := s.GetJobPods(namespace, name)
	if err == nil && len(pods) > 0 {
		// Get logs from pods
		debugInfo.Logs = s.getLogsFromPods(namespace, pods)

		// Convert to pod summaries
		for _, pod := range pods {
			debugInfo.Pods = append(debugInfo.Pods, PodSummary{
				Name:   pod.Name,
				Status: string(pod.Status.Phase),
				Node:   pod.Spec.NodeName,
			})
		}
	}

	// Get events
	events, err := s.getJobEvents(namespace, name)
	if err == nil {
		debugInfo.Events = events
	}

	return debugInfo, nil
}

// GetJobByUUID finds a job by its UUID using label selectors
func (s *JobDebugService) GetJobByUUID(uuid, namespace string) (*v1.Job, error) {
	ctx := context.Background()

	// Build label selector for UUID
	labelSelector := labels.Set{
		"job-uuid": uuid,
	}.AsSelector().String()

	// If no namespace specified, search all namespaces
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	// List jobs with the UUID label
	jobs, err := s.clientset.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}

	if len(jobs.Items) == 0 {
		// Try searching by annotation if label not found
		allJobs, err := s.clientset.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to list all jobs: %w", err)
		}

		for _, job := range allJobs.Items {
			if job.Annotations["job-uuid"] == uuid || job.Annotations["uuid"] == uuid {
				return &job, nil
			}
		}

		return nil, fmt.Errorf("job with UUID %s not found", uuid)
	}

	return &jobs.Items[0], nil
}

// GetJobTraces extracts trace information from job annotations
func (s *JobDebugService) GetJobTraces(namespace, name string) (*TraceInfo, error) {
	ctx := context.Background()

	job, err := s.clientset.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	return s.extractTraceInfo(job), nil
}

// GetJobErrors returns error information for a job
func (s *JobDebugService) GetJobErrors(namespace, name string) (*ErrorInfo, error) {
	ctx := context.Background()

	job, err := s.clientset.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	return s.getJobErrors(job)
}

// GetJobSandboxLogs returns sandbox logs for a job
func (s *JobDebugService) GetJobSandboxLogs(namespace, name string) (*LogInfo, error) {
	pods, err := s.GetJobPods(namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get job pods: %w", err)
	}

	return s.getLogsFromPods(namespace, pods), nil
}

// GetJobPods returns all pods associated with a job
func (s *JobDebugService) GetJobPods(namespace, name string) ([]corev1.Pod, error) {
	ctx := context.Background()

	// Get the job to extract selector
	job, err := s.clientset.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	// Use job's selector to find pods
	selector, err := metav1.LabelSelectorAsSelector(job.Spec.Selector)
	if err != nil {
		return nil, fmt.Errorf("failed to parse job selector: %w", err)
	}

	// List pods with the job's selector
	pods, err := s.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	return pods.Items, nil
}

// Helper functions

func (s *JobDebugService) getJobSummary(job *v1.Job) *JobSummary {
	status := "Unknown"
	if job.Status.Succeeded > 0 {
		status = "Succeeded"
	} else if job.Status.Failed > 0 {
		status = "Failed"
	} else if job.Status.Active > 0 {
		status = "Active"
	}

	uuid := ""
	if val, ok := job.Labels["job-uuid"]; ok {
		uuid = val
	} else if val, ok := job.Annotations["job-uuid"]; ok {
		uuid = val
	}

	return &JobSummary{
		Name:      job.Name,
		Namespace: job.Namespace,
		UUID:      uuid,
		Status:    status,
		Labels:    job.Labels,
	}
}

func (s *JobDebugService) extractJobError(job *v1.Job) *JobError {
	// Check if job has pre-categorized error in annotations
	annotations := job.Annotations
	if annotations == nil {
		return nil
	}

	// Look for JobError annotations (these would be set by your error categorization service)
	jobError := &JobError{}
	hasError := false

	if val, ok := annotations["job.error.category"]; ok {
		jobError.Category = val
		hasError = true
	}
	if val, ok := annotations["job.error.code"]; ok {
		jobError.ErrorCode = val
		hasError = true
	}
	if val, ok := annotations["job.error.description"]; ok {
		jobError.Description = val
		hasError = true
	}
	if val, ok := annotations["job.error.timestamp"]; ok {
		jobError.Timestamp = val
		hasError = true
	}
	if val, ok := annotations["job.error.severity"]; ok {
		jobError.Severity = val
		hasError = true
	}

	// If no pre-categorized error, try to infer from job status
	if !hasError && job.Status.Failed > 0 {
		jobError.Category = "JobFailure"
		jobError.Severity = "High"
		jobError.Timestamp = time.Now().Format(time.RFC3339)
		
		// Check job conditions for more details
		for _, condition := range job.Status.Conditions {
			if condition.Type == v1.JobFailed && condition.Status == corev1.ConditionTrue {
				jobError.ErrorCode = condition.Reason
				jobError.Description = condition.Message
				hasError = true
				break
			}
		}
	}

	if hasError {
		return jobError
	}
	return nil
}

func (s *JobDebugService) extractTraceInfo(job *v1.Job) *TraceInfo {
	trace := &TraceInfo{}

	// Check common annotation keys for Datadog traces
	annotations := job.Annotations
	if annotations == nil {
		return trace
	}

	// Common Datadog annotation keys
	if val, ok := annotations["datadog.trace.url"]; ok {
		trace.DatadogURL = val
	}
	if val, ok := annotations["datadog.trace.id"]; ok {
		trace.TraceID = val
	}
	if val, ok := annotations["datadog.span.id"]; ok {
		trace.SpanID = val
	}
	if val, ok := annotations["dd.trace.link"]; ok {
		trace.TraceLink = val
	}

	// Sandbox link might be stored here too
	if val, ok := annotations["sandbox.url"]; ok {
		trace.TraceLink = val
	}

	return trace
}

func (s *JobDebugService) getJobErrors(job *v1.Job) (*ErrorInfo, error) {
	errorInfo := &ErrorInfo{
		Type: "JobFailure",
	}

	// Check job conditions
	for _, condition := range job.Status.Conditions {
		if condition.Type == v1.JobFailed && condition.Status == corev1.ConditionTrue {
			errorInfo.Reason = condition.Reason
			errorInfo.Message = condition.Message
			errorInfo.Timestamp = condition.LastTransitionTime.String()
			break
		}
	}

	// Get pod errors
	pods, err := s.GetJobPods(job.Namespace, job.Name)
	if err == nil {
		for _, pod := range pods {
			if pod.Status.Phase == corev1.PodFailed {
				for _, containerStatus := range pod.Status.ContainerStatuses {
					if containerStatus.State.Terminated != nil && containerStatus.State.Terminated.ExitCode != 0 {
						errorInfo.PodErrors = append(errorInfo.PodErrors, PodError{
							PodName:   pod.Name,
							Container: containerStatus.Name,
							Reason:    containerStatus.State.Terminated.Reason,
							Message:   containerStatus.State.Terminated.Message,
						})
					}
				}
			}
		}
	}

	return errorInfo, nil
}

func (s *JobDebugService) getLogsFromPods(namespace string, pods []corev1.Pod) *LogInfo {
	logInfo := &LogInfo{
		Containers: make(map[string]string),
		LogFiles:   make(map[string]string),
	}

	// Extract sandbox URL/path from pod annotations if available
	for _, pod := range pods {
		if url, ok := pod.Annotations["sandbox.url"]; ok {
			logInfo.SandboxURL = url
		}
		if path, ok := pod.Annotations["sandbox.path"]; ok {
			logInfo.SandboxPath = path
			// List expected log files
			logInfo.LogFiles["stdout"] = "std.out"
			logInfo.LogFiles["stderr"] = "std.err"
			logInfo.LogFiles["decout"] = "decout"
			logInfo.LogFiles["decerr"] = "decerr"
		}
		if logInfo.SandboxURL != "" || logInfo.SandboxPath != "" {
			break
		}
	}

	// Note: Actual log retrieval would require the podLogEventService
	// This is a placeholder showing where logs would be collected
	for _, pod := range pods {
		for _, container := range pod.Spec.Containers {
			key := fmt.Sprintf("%s/%s", pod.Name, container.Name)
			logInfo.Containers[key] = fmt.Sprintf("Use /namespaces/%s/pods/%s/logs?container=%s to retrieve logs",
				namespace, pod.Name, container.Name)
		}
	}

	return logInfo
}

func (s *JobDebugService) getJobEvents(namespace, name string) ([]string, error) {
	ctx := context.Background()

	// Get events for the job
	events, err := s.clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Job", name),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	var eventMessages []string
	for _, event := range events.Items {
		eventMessages = append(eventMessages, fmt.Sprintf("[%s] %s: %s",
			event.Type, event.Reason, event.Message))
	}

	return eventMessages, nil
}

// ReadSandboxLogFile reads a specific log file from the sandbox directory
func (s *JobDebugService) ReadSandboxLogFile(sandboxPath, logFile string, startLine, numLines string) (string, error) {
	// Validate inputs
	start, err := strconv.Atoi(startLine)
	if err != nil {
		start = 0
	}
	
	lines, err := strconv.Atoi(numLines)
	if err != nil {
		lines = 1000
	}

	// Construct full file path
	fullPath := filepath.Join(sandboxPath, logFile)
	
	// Security check - ensure we're not accessing files outside sandbox
	if !strings.HasPrefix(fullPath, sandboxPath) {
		return "", fmt.Errorf("invalid file path - must be within sandbox directory")
	}

	// Open the file
	file, err := os.Open(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Read file line by line
	scanner := bufio.NewScanner(file)
	var result strings.Builder
	lineNum := 0
	capturedLines := 0

	for scanner.Scan() {
		if lineNum >= start && capturedLines < lines {
			result.WriteString(fmt.Sprintf("%d: %s\n", lineNum+1, scanner.Text()))
			capturedLines++
		}
		lineNum++
		
		if capturedLines >= lines {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	return result.String(), nil
}

// AnalyzeSandboxLogs analyzes sandbox logs for errors and important information
func (s *JobDebugService) AnalyzeSandboxLogs(sandboxPath string) (*SandboxAnalysis, error) {
	analysis := &SandboxAnalysis{
		ErrorLines:   []LogLine{},
		WarningLines: []LogLine{},
		KeyEvents:    []LogLine{},
	}

	// Common log files to analyze
	logFiles := []string{"std.out", "std.err", "decout", "decerr"}
	
	for _, logFile := range logFiles {
		fullPath := filepath.Join(sandboxPath, logFile)
		
		file, err := os.Open(fullPath)
		if err != nil {
			continue // Skip if file doesn't exist
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 0

		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			lowerLine := strings.ToLower(line)

			// Check for errors
			if strings.Contains(lowerLine, "error") || strings.Contains(lowerLine, "exception") || 
			   strings.Contains(lowerLine, "failed") || strings.Contains(lowerLine, "fatal") {
				analysis.ErrorLines = append(analysis.ErrorLines, LogLine{
					File:       logFile,
					LineNumber: lineNum,
					Content:    line,
				})
			}

			// Check for warnings
			if strings.Contains(lowerLine, "warning") || strings.Contains(lowerLine, "warn") {
				analysis.WarningLines = append(analysis.WarningLines, LogLine{
					File:       logFile,
					LineNumber: lineNum,
					Content:    line,
				})
			}

			// Check for key events (you can customize these patterns)
			if strings.Contains(lowerLine, "starting") || strings.Contains(lowerLine, "completed") ||
			   strings.Contains(lowerLine, "finished") || strings.Contains(lowerLine, "terminated") {
				analysis.KeyEvents = append(analysis.KeyEvents, LogLine{
					File:       logFile,
					LineNumber: lineNum,
					Content:    line,
				})
			}
		}
	}

	return analysis, nil
}

type SandboxAnalysis struct {
	ErrorLines   []LogLine `json:"errorLines"`
	WarningLines []LogLine `json:"warningLines"`
	KeyEvents    []LogLine `json:"keyEvents"`
}

type LogLine struct {
	File       string `json:"file"`
	LineNumber int    `json:"lineNumber"`
	Content    string `json:"content"`
}

type JobError struct {
	Category    string `json:"category"`
	ErrorCode   string `json:"errorCode"`
	Description string `json:"description"`
	Timestamp   string `json:"timestamp"`
	Severity    string `json:"severity"`
}
