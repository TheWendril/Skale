package controller

import (
	"context"
	"fmt"
	"time"

	skalepbv1 "github.com/TheWendril/Skale/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type SkaleReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	MetricsClient *metrics.Clientset
}

func (r *SkaleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var skale skalepbv1.Skale
	if err := r.Get(ctx, req.NamespacedName, &skale); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	targetRef := skale.Spec.ScaleTargetRef

	if targetRef.Kind != "Deployment" {
		log.Info("Somente Deployment é suportado no momento")
		return ctrl.Result{}, nil
	}

	var deploy appsv1.Deployment
	if err := r.Get(ctx, types.NamespacedName{Name: targetRef.Name, Namespace: req.Namespace}, &deploy); err != nil {
		if errors.IsNotFound(err) {
			log.Info("Deployment não encontrado")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	var podList corev1.PodList
	if err := r.List(ctx, &podList, client.InNamespace(req.Namespace), client.MatchingLabels(deploy.Spec.Selector.MatchLabels)); err != nil {
		return ctrl.Result{}, err
	}

	podMetricsList, err := r.MetricsClient.MetricsV1beta1().PodMetricses(req.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Error(err, "Erro ao listar métricas dos pods")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}
	log.Info("Coletando métricas dos pods", "quantidade", len(podMetricsList.Items))

	var (
		totalUsageCPU, totalLimitCPU       int64
		totalUsageMemory, totalLimitMemory int64
	)

	for _, pod := range podList.Items {
		for _, c := range pod.Spec.Containers {
			if cpu := c.Resources.Limits.Cpu(); cpu != nil {
				totalLimitCPU += cpu.MilliValue()
			}
			if mem := c.Resources.Limits.Memory(); mem != nil {
				totalLimitMemory += mem.Value()
			}
		}
	}

	validPods := make(map[string]bool)
	for _, pod := range podList.Items {
		validPods[pod.Name] = true
	}

	for _, m := range podMetricsList.Items {
		if !validPods[m.Name] {
			continue
		}
		for _, c := range m.Containers {
			totalUsageCPU += c.Usage.Cpu().MilliValue()
			totalUsageMemory += c.Usage.Memory().Value()
			log.Info("Uso de recursos do container",
				"pod", m.Name,
				"container", c.Name,
				"cpu (millicores)", c.Usage.Cpu().MilliValue(),
				"mem (MiB)", float64(c.Usage.Memory().Value())/1024.0/1024.0,
			)
		}
	}

	if deploy.Spec.Replicas == nil || (totalLimitCPU == 0 && totalLimitMemory == 0) {
		log.Info("Sem limites ou réplicas definidas")
		return ctrl.Result{}, nil
	}

	var (
		targetCPUUtilization    int32 = 80
		targetMemoryUtilization int32 = 80
	)

	for _, metric := range skale.Spec.Metrics {
		if metric.Type == "Resource" && metric.Resource.TargetAverageUtilization != nil {
			switch metric.Resource.Name {
			case "cpu":
				targetCPUUtilization = *metric.Resource.TargetAverageUtilization
			case "memory":
				targetMemoryUtilization = *metric.Resource.TargetAverageUtilization
			}
		}
	}

	cpuRatio := ratio(totalUsageCPU, totalLimitCPU)
	memRatio := ratio(totalUsageMemory, totalLimitMemory)

	cpuTarget := float64(targetCPUUtilization) / 100.0
	memTarget := float64(targetMemoryUtilization) / 100.0

	percentUsed := max(cpuRatio, memRatio)
	target := max(cpuTarget, memTarget)

	currentReplicas := *deploy.Spec.Replicas
	minReplicas := skale.Spec.MinReplicas
	maxReplicas := skale.Spec.MaxReplicas

	log.Info("Uso atual de recursos",
		"CPU", fmt.Sprintf("%.2f%%", cpuRatio*100),
		"Memory", fmt.Sprintf("%.2f%%", memRatio*100),
	)

	log.Info("Escalonamento - dados atuais",
		"totalUsageCPU", totalUsageCPU,
		"totalLimitCPU", totalLimitCPU,
		"totalUsageMemory", totalUsageMemory,
		"totalLimitMemory", totalLimitMemory,
		"percentUsed", percentUsed,
		"targetCPU", targetCPUUtilization,
		"targetMemory", targetMemoryUtilization,
		"targetRatio", fmt.Sprintf("%.2f", target),
		"currentReplicas", currentReplicas,
		"minReplicas", minReplicas,
		"maxReplicas", maxReplicas,
	)

	if percentUsed > target && currentReplicas < maxReplicas {
		*deploy.Spec.Replicas++
		log.Info("Escalando UP", "replicas", *deploy.Spec.Replicas)
		if err := r.Update(ctx, &deploy); err != nil {
			return ctrl.Result{}, err
		}
	} else if percentUsed < target*0.6 && currentReplicas > minReplicas {
		*deploy.Spec.Replicas--
		log.Info("Escalando DOWN", "replicas", *deploy.Spec.Replicas)
		if err := r.Update(ctx, &deploy); err != nil {
			return ctrl.Result{}, err
		}
	}

	if currentReplicas < minReplicas {
		*deploy.Spec.Replicas = minReplicas
		log.Info("Aumentando para o mínimo de réplicas", "replicas", minReplicas)
		if err := r.Update(ctx, &deploy); err != nil {
			return ctrl.Result{}, err
		}
	}

	if currentReplicas > maxReplicas {
		*deploy.Spec.Replicas = maxReplicas
		log.Info("Aumentando para o máximmo de réplicas", "replicas", minReplicas)
		if err := r.Update(ctx, &deploy); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

func (r *SkaleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&skalepbv1.Skale{}).
		Complete(r)
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func ratio(used, limit int64) float64 {
	if limit == 0 {
		return 0
	}
	return float64(used) / float64(limit)
}
