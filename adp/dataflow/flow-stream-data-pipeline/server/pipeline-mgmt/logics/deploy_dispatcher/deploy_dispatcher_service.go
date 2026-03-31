package deploy_dispatcher

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/kubernetes"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"flow-stream-data-pipeline/pipeline-mgmt/interfaces"
)

// use to create, update, delete kubernetes resource by kubernetes api
type deployDispatcherService struct {
	// kubernetes client
	k8sCli    *kubernetes.KubernetesClient
	namespace string
}

var ref = metav1.OwnerReference{
	APIVersion: "apps/v1",
	Kind:       "Deployment",
}

func NewDeployDispatcherService(namespace string) (interfaces.DeployDispatcherService, error) {
	k8sCli, err := kubernetes.NewKubernetesClient(namespace)
	if err != nil {
		return nil, err
	}
	return &deployDispatcherService{
		k8sCli:    k8sCli,
		namespace: namespace,
	}, nil
}

// create kubernetes Deploy
func (d *deployDispatcherService) CreateDeploy(ctx context.Context, pipelineId string, k8scfg *interfaces.KubernetesCfg, serviceInfo *interfaces.ServiceInfo) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "create pipeline deploy")
	span.SetAttributes(
		attr.Key("pipeline_id").String(pipelineId),
		attr.Key("namespace").String(d.namespace))
	defer span.End()

	if err := d.setRef(serviceInfo.ManagerDeployName); err != nil {
		// 设置 span 为失败状态
		span.SetStatus(codes.Error, "failed to setRef")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("faield to setRef, error: %v", err))

		return err
	}
	_, err := d.k8sCli.GetDeploy(ctx, d.getDeployName(pipelineId, serviceInfo.ServiceName))

	ref.Name = serviceInfo.ManagerDeployName
	deployName := d.getDeployName(pipelineId, serviceInfo.ServiceName)

	span.SetAttributes(attr.Key("deploy_name").String(deployName))

	if kubernetes.ResourceNotFound(err) {
		container := d.newWorkerContainer(k8scfg)
		_, err = d.k8sCli.CreateDeploy(
			ctx,
			k8scfg.VolumeMounts,
			kubernetes.WithName(d.getDeployName(pipelineId, serviceInfo.ServiceName)),
			kubernetes.WithLabel(k8scfg.Label),
			kubernetes.WithMatchLabel(k8scfg.Label),
			kubernetes.WithPodLabel(k8scfg.Label),
			kubernetes.WithReplicas(1),
			kubernetes.WithStrategy(appsv1.DeploymentStrategy{Type: appsv1.RecreateDeploymentStrategyType}),
			kubernetes.WithContainer(container),
			kubernetes.WithServiceAccount(serviceInfo.ServiceAccount),
			kubernetes.WithOwnerReference([]metav1.OwnerReference{ref}),
		)
		if err != nil {
			// 设置 span 为失败状态
			span.SetStatus(codes.Error, "create deploy failed")
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("failed to create deploy[%s], error: %v", deployName, err))

			return err
		}
	}
	if err != nil {
		// 设置 span 为失败状态
		span.SetStatus(codes.Error, "get deploy failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("failed to get deploy [%s], error: %v", deployName, err))
	} else {
		span.SetStatus(codes.Ok, "")
	}
	return err
}

// update kubernetes Deploy
func (d *deployDispatcherService) UpdateDeploy(ctx context.Context, pipelineId string, k8scfg *interfaces.KubernetesCfg, serviceInfo *interfaces.ServiceInfo) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "update pipeline deploy")
	span.SetAttributes(
		attr.Key("pipeline_id").String(pipelineId),
		attr.Key("namespace").String(d.namespace))
	defer span.End()

	if err := d.setRef(serviceInfo.ManagerDeployName); err != nil {
		// 设置 span 为失败状态
		span.SetStatus(codes.Error, "failed to setRef")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("faield to setRef, error: %v", err))

		return err
	}

	deployName := d.getDeployName(pipelineId, serviceInfo.ServiceName)
	span.SetAttributes(attr.Key("deploy_name").String(deployName))

	ds, err := d.k8sCli.GetDeploy(ctx, deployName)
	if err != nil && kubernetes.ResourceNotFound(err) {
		span.SetStatus(codes.Error, "get deploy failed")
		o11y.Error(ctx, fmt.Sprintf("failed to get deploy[%s], error: %v", deployName, err))

		return err
	}

	ref.Name = serviceInfo.ManagerDeployName
	ds.Labels = k8scfg.Label
	ds.Spec.Template.Spec.Containers = []corev1.Container{d.newWorkerContainer(k8scfg)}
	ds.OwnerReferences = []metav1.OwnerReference{ref}
	// 使用配置中的 volumes
	ds.Spec.Template.Spec.Volumes = k8scfg.Volumes

	_, err = d.k8sCli.UpdateDeploy(ctx, ds)
	if err != nil {
		span.SetStatus(codes.Error, "update deploy failed")
		o11y.Error(ctx, fmt.Sprintf("failed to update deploy [%s], error: %v", deployName, err))
	} else {
		span.SetStatus(codes.Ok, "")
	}
	return err
}

// clean up the k8s deploy when unbinding deploy or existing
func (d *deployDispatcherService) DeleteDeploy(ctx context.Context, pipelineId string, serviceInfo *interfaces.ServiceInfo) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "delete pipeline deploy")
	span.SetAttributes(
		attr.Key("pipeline_id").String(pipelineId),
		attr.Key("namespace").String(d.namespace))
	defer span.End()

	deployName := d.getDeployName(pipelineId, serviceInfo.ServiceName)

	span.SetAttributes(attr.Key("deploy_name").String(deployName))

	err := d.k8sCli.DeleteDeploy(ctx, deployName)
	if err != nil {
		span.SetStatus(codes.Error, "delete deploy failed")
		o11y.Error(ctx, fmt.Sprintf("failed to delete deploy [%s], error: %v", deployName, err))
	} else {
		span.SetStatus(codes.Ok, "")
	}
	return err
}

// 根据名字获取
func (d *deployDispatcherService) GetDeploy(ctx context.Context, pipelineId string, serviceInfo *interfaces.ServiceInfo) (deploy *appsv1.Deployment, err error) {
	return d.k8sCli.GetDeploy(ctx, d.getDeployName(pipelineId, serviceInfo.ServiceName))
}

// 根据 opts 获取 deploy 列表
func (d *deployDispatcherService) ListDeploy(ctx context.Context, opts metav1.ListOptions) (deployList *appsv1.DeploymentList, err error) {
	return d.k8sCli.ListDeploy(ctx, opts)
}

func (d *deployDispatcherService) ApplyData(ctx context.Context, action interfaces.ApplyAction, data interfaces.ApplyData) error {
	return nil
}

func (d *deployDispatcherService) getDeployName(pipelineId string, serviceName string) string {
	return fmt.Sprintf("%s-%s", serviceName, pipelineId)
}

func (d *deployDispatcherService) setRef(managerDeployName string) error {
	if ref.UID == "" {
		dep, err := d.k8sCli.GetDeploy(context.Background(), managerDeployName)

		if err != nil {
			return err
		}
		ref.UID = dep.GetUID()
	}
	return nil
}

// new container
func (d *deployDispatcherService) newWorkerContainer(cfg *interfaces.KubernetesCfg) corev1.Container {
	container := kubernetes.NewContainer(
		kubernetes.ContainerWithImage(cfg.Image),
		kubernetes.ContainerWithCommand(cfg.ContainerCmd, cfg.ContainerArgs),
		kubernetes.ContainerWithWorkDir(cfg.WorkingDir),
		kubernetes.ContainerWithResource(cfg.Resource),
		kubernetes.ContainerWithVolumeMount(cfg.VolumeMounts),
	)
	return *container
}
