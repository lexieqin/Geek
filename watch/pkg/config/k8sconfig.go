package config

import (
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type K8sConfig struct {
	*rest.Config
	*kubernetes.Clientset
	e error
}

func NewK8sConfig() *K8sConfig {
	return &K8sConfig{}
}

// 初始化k8s配置
func (k *K8sConfig) InitRestConfig() *K8sConfig {
	config, _ := clientcmd.BuildConfigFromFlags("", "/Users/qinqin/.kube/config")
	k.Config = config

	return k
}

func (k *K8sConfig) Error() error {
	return k.e
}

// 初始化clientSet客户端
func (k *K8sConfig) InitClientSet() *kubernetes.Clientset {
	if k.Config == nil {
		k.e = errors.Wrap(errors.New("k8s config is nil"), "init k8s client failed")
		return nil
	}

	clientSet, err := kubernetes.NewForConfig(k.Config)
	if err != nil {
		k.e = errors.Wrap(err, "init k8s clientSet failed")
		return nil
	}
	return clientSet
}
