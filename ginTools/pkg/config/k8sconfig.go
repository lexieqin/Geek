package config

import (
	"fmt"
	"path/filepath"

	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type K8sConfig struct {
	*rest.Config
	*kubernetes.Clientset
	*dynamic.DynamicClient
	meta.RESTMapper
	informers.SharedInformerFactory
	e error
}

func NewK8sConfig() *K8sConfig {
	return &K8sConfig{}
}

// 初始化k8s配置 - 优先使用 in-cluster config，失败则使用 kubeconfig
func (k *K8sConfig) InitRestConfig(optfuncs ...K8sConfigOptionFunc) *K8sConfig {
	// First try in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		// If in-cluster fails, fall back to kubeconfig
		fmt.Println("Not running in cluster, using kubeconfig")
		kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			k.e = errors.Wrap(err, "failed to initialize k8s config (both in-cluster and kubeconfig)")
			return k
		}
		fmt.Printf("Using kubeconfig from: %s\n", kubeconfig)
	} else {
		fmt.Println("Running in cluster, using in-cluster config")
	}
	
	k.Config = config
	for _, optfunc := range optfuncs {
		optfunc(k)
	}
	return k
}

func (k *K8sConfig) InitConfigInCluster() *K8sConfig {
	// 加载 in-cluster 配置
	config, err := rest.InClusterConfig()
	if err != nil {
		k.e = errors.Wrap(errors.New("k8s config is nil"), "init k8s client failed")
	}
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

// 初始化动态客户端
func (k *K8sConfig) InitDynamicClient() *dynamic.DynamicClient {
	if k.Config == nil {
		k.e = errors.Wrap(errors.New("k8s config is nil"), "init k8s client failed")
		return nil
	}

	dynamicClient, err := dynamic.NewForConfig(k.Config)
	if err != nil {
		k.e = errors.Wrap(err, "init k8s dynamicClient failed")
		return nil
	}
	return dynamicClient
}

// 获取  所有api groupresource
func (k *K8sConfig) InitRestMapper() meta.RESTMapper {
	gr, err := restmapper.GetAPIGroupResources(k.InitClientSet().Discovery())
	if err != nil {
		k.e = errors.Wrap(err, "init k8s dynamicClient failed")
	}
	mapper := restmapper.NewDiscoveryRESTMapper(gr)

	return mapper
}

func (k *K8sConfig) InitInformer() informers.SharedInformerFactory {
	fact := informers.NewSharedInformerFactory(k.InitClientSet(), 0) //创建通用informer工厂

	informer := fact.Core().V1().Pods()
	informer.Informer().AddEventHandler(&cache.ResourceEventHandlerFuncs{})

	ch := make(chan struct{})
	fact.Start(ch)
	fact.WaitForCacheSync(ch)

	return fact
}

type K8sConfigOptionFunc func(k *K8sConfig)

func WithQps(qps float32) K8sConfigOptionFunc {
	return func(k *K8sConfig) {
		if k.Config != nil {
			k.QPS = qps
		}
	}
}

func WithBurst(b int) K8sConfigOptionFunc {
	return func(k *K8sConfig) {
		if k.Config != nil {
			k.Burst = b
		}
	}
}
