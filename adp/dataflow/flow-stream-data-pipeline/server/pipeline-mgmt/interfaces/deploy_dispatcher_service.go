package interfaces

import (
	"context"

	"github.com/kweaver-ai/kweaver-go-lib/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KubernetesCfg struct {
	Namespace     string
	Image         kubernetes.Image           `json:"image"`
	Resource      kubernetes.ResourceRequire `json:"client"`
	Label         map[string]string          `json:"label"`
	ContainerCmd  []string
	ContainerArgs []string
	WorkingDir    string
	VolumeMounts  []corev1.VolumeMount
	Volumes       []corev1.Volume
}

type ApplyAction string
type ApplyData interface{}

type ServiceInfo struct {
	ManagerDeployName string
	ServiceName       string
	ServiceAccount    string
}

//go:generate mockgen -source ../interfaces/deploy_dispatcher_service.go -destination ../interfaces/mock/mock_deploy_dispatcher_service.go
type DeployDispatcherService interface {
	CreateDeploy(ctx context.Context, pipelineID string, k8scfg *KubernetesCfg, serviceInfo *ServiceInfo) error
	UpdateDeploy(ctx context.Context, pipelineID string, k8scfg *KubernetesCfg, serviceInfo *ServiceInfo) error
	DeleteDeploy(ctx context.Context, pipelineID string, serviceInfo *ServiceInfo) error
	GetDeploy(ctx context.Context, pipelineID string, serviceInfo *ServiceInfo) (deploy *appsv1.Deployment, err error)
	ListDeploy(ctx context.Context, opts metav1.ListOptions) (deployList *appsv1.DeploymentList, err error)
	ApplyData(context.Context, ApplyAction, ApplyData) error
}
