package cmd

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type GitlabGroupDescriptor struct {
	Name string `yaml:"name"`
	Id   int    `yaml:"id"`
}
type GitGroupDescriptor struct {
	GroupByUrls []string                `yaml:"groupByUrls"`
	GroupByIds  []GitlabGroupDescriptor `yaml:"groupByIds"`
}

type GitLabSearchInfo struct {
	GroupDescriptor GitGroupDescriptor `yaml:"groupDescriptor"`
	BaseUrl         string             `yaml:"baseUrl"`
	Page            int                `yaml:"page"`
	RowPerPage      int                `yaml:"rowPerPage"`
}

type AppConfig struct {
	Port         int                      `yaml:"port"`
	EndPoints    []map[string]interface{} `yaml:"endPoints"`
	GitlabSearch GitLabSearchInfo         `yaml:"gitlabSearch"`
}

func buildKubeClient() (*kubernetes.Clientset, error) {
	var errConfig error
	var restConfig *rest.Config
	var kClientSet *kubernetes.Clientset
	if kClientCmdParams.inCluster == "false" {
		hDir := os.Getenv("HOME")
		fPath := filepath.Join(hDir, ".kube", "config")
		restConfig, errConfig = clientcmd.BuildConfigFromFlags("", fPath)

	} else {
		restConfig, errConfig = clientcmd.BuildConfigFromFlags("", "")
	}
	if errConfig != nil {
		Logger.Fatal("error while getting config for kubernetes", zap.Error(errConfig))
		return nil, errConfig

	} else {

		Logger.Info("kube config is created", zap.String("config", restConfig.Host))
		kClientSet, errConfig = kubernetes.NewForConfig(restConfig)
		if errConfig != nil {
			Logger.Fatal("error while creating clientset", zap.Error(errConfig))
			return nil, errConfig
		}
		Logger.Info("successfully create clientset")

		return kClientSet, nil
	}

}
