package services

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// PodLogEventService handles operations related to pod logs and events
type PodLogEventService struct {
	client *kubernetes.Clientset
}

// NewPodLogEventService creates a new instance of PodLogEventService
func NewPodLogEventService(client *kubernetes.Clientset) *PodLogEventService {
	return &PodLogEventService{client: client}
}

// GetLogs returns a request to retrieve logs from a specific pod
// Parameters:
//   - ns: namespace where the pod is located
//   - podName: name of the pod
//   - tailLine: number of lines to show from the end of the logs
//   - containerName: optional container name (if empty, uses the first container)
func (s *PodLogEventService) GetLogs(ns, podName string, tailLine int64, containerName string) *rest.Request {
	options := &v1.PodLogOptions{
		Follow:    false,
		TailLines: &tailLine,
	}

	// Only set the container name if it's provided
	if containerName != "" {
		options.Container = containerName
	}

	return s.client.CoreV1().Pods(ns).GetLogs(podName, options)
}

// PodEvent represents a filtered Kubernetes event
type PodEvent struct {
	Type      string    `json:"type"`
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// GetEvents retrieves events related to a specific pod
// Parameters:
//   - ns: namespace where the pod is located
//   - podName: name of the pod
//   - eventType: optional filter for event type (e.g., "Warning", "Normal")
func (s *PodLogEventService) GetEvents(ns, podName string, eventType string) ([]PodEvent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	events, err := s.client.CoreV1().Events(ns).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Pod", podName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	var podEvents []PodEvent
	for _, event := range events.Items {
		// If eventType is specified, filter by that type
		if eventType != "" && event.Type != eventType {
			continue
		}

		podEvents = append(podEvents, PodEvent{
			Type:      event.Type,
			Reason:    event.Reason,
			Message:   event.Message,
			Timestamp: event.CreationTimestamp.Time,
		})
	}

	return podEvents, nil
}

// Pod represents a simplified Kubernetes pod structure for API responses
type Pod struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Status     string            `json:"status"`
	Phase      v1.PodPhase       `json:"phase"`
	Conditions []v1.PodCondition `json:"conditions"`
	StartTime  *metav1.Time      `json:"startTime"`
	IP         string            `json:"ip"`
	NodeName   string            `json:"nodeName"`
	Labels     map[string]string `json:"labels"`
}

// ListPods returns all pods in the specified namespace
func (s *PodLogEventService) ListPods(ns string) ([]Pod, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pods, err := s.client.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	result := make([]Pod, 0, len(pods.Items))
	for _, pod := range pods.Items {
		result = append(result, convertToPod(&pod))
	}

	return result, nil
}

// GetPod retrieves a specific pod by name and namespace
func (s *PodLogEventService) GetPod(ns, podName string) (*Pod, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pod, err := s.client.CoreV1().Pods(ns).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod %s in namespace %s: %w", podName, ns, err)
	}

	result := convertToPod(pod)
	return &result, nil
}

// convertToPod creates a simplified Pod representation from a Kubernetes Pod
func convertToPod(pod *v1.Pod) Pod {
	var status string
	if pod.Status.Phase == v1.PodRunning {
		status = "Running"
		for _, condition := range pod.Status.Conditions {
			if condition.Status != v1.ConditionTrue {
				status = "Not Ready"
				break
			}
		}
	} else {
		status = string(pod.Status.Phase)
	}

	return Pod{
		Name:       pod.Name,
		Namespace:  pod.Namespace,
		Status:     status,
		Phase:      pod.Status.Phase,
		Conditions: pod.Status.Conditions,
		StartTime:  pod.Status.StartTime,
		IP:         pod.Status.PodIP,
		NodeName:   pod.Spec.NodeName,
		Labels:     pod.Labels,
	}
}
