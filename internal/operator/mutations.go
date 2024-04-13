package operator

import (
	"context"

	"github.com/cloudnative-pg/cnpg-i-machinery/pkg/pluginhelper"
	"github.com/cloudnative-pg/cnpg-i/pkg/operator"
	corev1 "k8s.io/api/core/v1"

	"github.com/dougkirkley/plugin-objstore-backup/pkg/metadata"
)

// MutateCluster is called to mutate a cluster with the defaulting webhook.
// This function is defaulting the "imagePullPolicy" plugin parameter
func (Operator) MutateCluster(
	_ context.Context,
	request *operator.OperatorMutateClusterRequest,
) (*operator.OperatorMutateClusterResult, error) {
	helper, err := pluginhelper.NewDataBuilder(metadata.Data.Name, request.Definition).Build()
	if err != nil {
		return nil, err
	}

	mutatedCluster := helper.GetCluster().DeepCopy()
	for i := range mutatedCluster.Spec.Plugins {
		if mutatedCluster.Spec.Plugins[i].Name != metadata.Data.Name {
			continue
		}

		if mutatedCluster.Spec.Plugins[i].Parameters == nil {
			mutatedCluster.Spec.Plugins[i].Parameters = make(map[string]string)
		}

		if _, ok := mutatedCluster.Spec.Plugins[i].Parameters[imagePullPolicyParameter]; !ok {
			mutatedCluster.Spec.Plugins[i].Parameters[imagePullPolicyParameter] = string(corev1.PullAlways)
		}
	}

	patch, err := helper.CreateClusterJSONPatch(*mutatedCluster)
	if err != nil {
		return nil, err
	}

	return &operator.OperatorMutateClusterResult{
		JsonPatch: patch,
	}, nil
}

// MutatePod is called to mutate a Pod before it will be created
func (Operator) MutatePod(
	_ context.Context,
	request *operator.OperatorMutatePodRequest,
) (*operator.OperatorMutatePodResult, error) {
	helper, err := pluginhelper.NewDataBuilder(metadata.Data.Name, request.ClusterDefinition).
		WithPod(request.PodDefinition).Build()
	if err != nil {
		return nil, err
	}

	mutatedPod := helper.GetPod().DeepCopy()
	helper.InjectPluginVolume(mutatedPod)

	// Inject sidecar
	if len(mutatedPod.Spec.Containers) > 0 {
		mutatedPod.Spec.Containers = append(
			mutatedPod.Spec.Containers,
			getSidecarContainer(mutatedPod, helper.Parameters))
	}

	// Inject backup volume
	if len(mutatedPod.Spec.Volumes) > 0 {
		mutatedPod.Spec.Volumes = append(
			mutatedPod.Spec.Volumes,
			getBackupVolume(helper.Parameters))
	}

	patch, err := helper.CreatePodJSONPatch(*mutatedPod)
	if err != nil {
		return nil, err
	}

	return &operator.OperatorMutatePodResult{
		JsonPatch: patch,
	}, nil
}
