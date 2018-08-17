/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package collectors

import (
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kube-state-metrics/pkg/constant"
	"k8s.io/kube-state-metrics/pkg/options"
	"k8s.io/kubernetes/pkg/apis/core/v1/helper"
)

var (
	descNodeLabelsName          = "kube_node_labels"
	descNodeLabelsHelp          = "Kubernetes labels converted to Prometheus labels."
	descNodeLabelsDefaultLabels = []string{"node"}

	descNodeInfo = prometheus.NewDesc(
		"kube_node_info",
		"Information about a cluster node.",
		append(descNodeLabelsDefaultLabels,
			"kernel_version",
			"os_image",
			"container_runtime_version",
			"kubelet_version",
			"kubeproxy_version",
			"provider_id"),
		nil,
	)
	descNodeCreated = prometheus.NewDesc(
		"kube_node_created",
		"Unix creation timestamp",
		descNodeLabelsDefaultLabels,
		nil,
	)
	descNodeLabels = prometheus.NewDesc(
		descNodeLabelsName,
		descNodeLabelsHelp,
		descNodeLabelsDefaultLabels,
		nil,
	)
	descNodeSpecUnschedulable = prometheus.NewDesc(
		"kube_node_spec_unschedulable",
		"Whether a node can schedule new pods.",
		descNodeLabelsDefaultLabels,
		nil,
	)
	descNodeSpecTaint = prometheus.NewDesc(
		"kube_node_spec_taint",
		"The taint of a cluster node.",
		append(descNodeLabelsDefaultLabels, "key", "value", "effect"),
		nil,
	)
	descNodeStatusCondition = prometheus.NewDesc(
		"kube_node_status_condition",
		"The condition of a cluster node.",
		append(descNodeLabelsDefaultLabels, "condition", "status"),
		nil,
	)
	descNodeStatusPhase = prometheus.NewDesc(
		"kube_node_status_phase",
		"The phase the node is currently in.",
		append(descNodeLabelsDefaultLabels, "phase"),
		nil,
	)
	descNodeStatusCapacity = prometheus.NewDesc(
		"kube_node_status_capacity",
		"The capacity for different resources of a node.",
		append(descNodeLabelsDefaultLabels, "resource", "unit"),
		nil,
	)
	descNodeStatusCapacityPods = prometheus.NewDesc(
		"kube_node_status_capacity_pods",
		"The total pod resources of the node.",
		descNodeLabelsDefaultLabels,
		nil,
	)
	descNodeStatusCapacityCPU = prometheus.NewDesc(
		"kube_node_status_capacity_cpu_cores",
		"The total CPU resources of the node.",
		descNodeLabelsDefaultLabels,
		nil,
	)
	descNodeStatusCapacityMemory = prometheus.NewDesc(
		"kube_node_status_capacity_memory_bytes",
		"The total memory resources of the node.",
		descNodeLabelsDefaultLabels,
		nil,
	)
	descNodeStatusAllocatable = prometheus.NewDesc(
		"kube_node_status_allocatable",
		"The allocatable for different resources of a node that are available for scheduling.",
		append(descNodeLabelsDefaultLabels, "resource", "unit"),
		nil,
	)
	descNodeStatusAllocatablePods = prometheus.NewDesc(
		"kube_node_status_allocatable_pods",
		"The pod resources of a node that are available for scheduling.",
		descNodeLabelsDefaultLabels,
		nil,
	)
	descNodeStatusAllocatableCPU = prometheus.NewDesc(
		"kube_node_status_allocatable_cpu_cores",
		"The CPU resources of a node that are available for scheduling.",
		descNodeLabelsDefaultLabels,
		nil,
	)
	descNodeStatusAllocatableMemory = prometheus.NewDesc(
		"kube_node_status_allocatable_memory_bytes",
		"The memory resources of a node that are available for scheduling.",
		descNodeLabelsDefaultLabels,
		nil,
	)
)

type NodeLister func() (v1.NodeList, error)

func (l NodeLister) List() (v1.NodeList, error) {
	return l()
}

func RegisterNodeCollector(registry prometheus.Registerer, client ClientSet, namespaces []string, opts *options.Options) {
	k8sclient := client.KubeClient.CoreV1().RESTClient()
	glog.Infof("collect node with %s", k8sclient.APIVersion())

	ninfs := NewSharedInformerList(k8sclient, "nodes", []string{metav1.NamespaceAll}, &v1.Node{})

	nodeLister := NodeLister(func() (machines v1.NodeList, err error) {
		for _, ninf := range *ninfs {
			for _, m := range ninf.GetStore().List() {
				machines.Items = append(machines.Items, *(m.(*v1.Node)))
			}
		}
		return machines, nil
	})

	registry.MustRegister(&nodeCollector{store: nodeLister, opts: opts})
	ninfs.Run(context.Background().Done())
}

type nodeStore interface {
	List() (v1.NodeList, error)
}

// nodeCollector collects metrics about all nodes in the cluster.
type nodeCollector struct {
	store nodeStore
	opts  *options.Options
}

// Describe implements the prometheus.Collector interface.
func (nc *nodeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- descNodeInfo
	ch <- descNodeCreated
	ch <- descNodeLabels
	ch <- descNodeSpecUnschedulable
	ch <- descNodeSpecTaint
	ch <- descNodeStatusCondition
	ch <- descNodeStatusPhase
	ch <- descNodeStatusCapacity
	ch <- descNodeStatusAllocatable

	if !nc.opts.DisableNodeNonGenericResourceMetrics {
		ch <- descNodeStatusCapacityCPU
		ch <- descNodeStatusCapacityMemory
		ch <- descNodeStatusCapacityPods
		ch <- descNodeStatusAllocatableCPU
		ch <- descNodeStatusAllocatableMemory
		ch <- descNodeStatusAllocatablePods
	}
}

// Collect implements the prometheus.Collector interface.
func (nc *nodeCollector) Collect(ch chan<- prometheus.Metric) {
	nodes, err := nc.store.List()
	if err != nil {
		ScrapeErrorTotalMetric.With(prometheus.Labels{"resource": "node"}).Inc()
		glog.Errorf("listing nodes failed: %s", err)
		return
	}
	ScrapeErrorTotalMetric.With(prometheus.Labels{"resource": "node"}).Add(0)

	ResourcesPerScrapeMetric.With(prometheus.Labels{"resource": "node"}).Observe(float64(len(nodes.Items)))
	for _, n := range nodes.Items {
		nc.collectNode(ch, n)
	}

	glog.V(4).Infof("collected %d nodes", len(nodes.Items))
}

func nodeLabelsDesc(labelKeys []string) *prometheus.Desc {
	return prometheus.NewDesc(
		descNodeLabelsName,
		descNodeLabelsHelp,
		append(descNodeLabelsDefaultLabels, labelKeys...),
		nil,
	)
}

func (nc *nodeCollector) collectNode(ch chan<- prometheus.Metric, n v1.Node) {
	addGauge := func(desc *prometheus.Desc, v float64, lv ...string) {
		lv = append([]string{n.Name}, lv...)
		ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, v, lv...)
	}
	// NOTE: the instrumentation API requires providing label values in order of declaration
	// in the metric descriptor. Be careful when making modifications.
	addGauge(descNodeInfo, 1,
		n.Status.NodeInfo.KernelVersion,
		n.Status.NodeInfo.OSImage,
		n.Status.NodeInfo.ContainerRuntimeVersion,
		n.Status.NodeInfo.KubeletVersion,
		n.Status.NodeInfo.KubeProxyVersion,
		n.Spec.ProviderID,
	)
	if !n.CreationTimestamp.IsZero() {
		addGauge(descNodeCreated, float64(n.CreationTimestamp.Unix()))
	}
	labelKeys, labelValues := kubeLabelsToPrometheusLabels(n.Labels)
	addGauge(nodeLabelsDesc(labelKeys), 1, labelValues...)

	addGauge(descNodeSpecUnschedulable, boolFloat64(n.Spec.Unschedulable))

	// Collect node taints
	for _, taint := range n.Spec.Taints {
		// Taints are applied to repel pods from nodes that do not have a corresponding
		// toleration.  Many node conditions are optionally reflected as taints
		// by the node controller in order to simplify scheduling constraints.
		addGauge(descNodeSpecTaint, 1, taint.Key, taint.Value, string(taint.Effect))
	}

	// Collect node conditions and while default to false.
	for _, c := range n.Status.Conditions {
		// This all-in-one metric family contains all conditions for extensibility.
		// Third party plugin may report customized condition for cluster node
		// (e.g. node-problem-detector), and Kubernetes may add new core
		// conditions in future.
		addConditionMetrics(ch, descNodeStatusCondition, c.Status, n.Name, string(c.Type))
	}

	// Set current phase to 1, others to 0 if it is set.
	if p := n.Status.Phase; p != "" {
		addGauge(descNodeStatusPhase, boolFloat64(p == v1.NodePending), string(v1.NodePending))
		addGauge(descNodeStatusPhase, boolFloat64(p == v1.NodeRunning), string(v1.NodeRunning))
		addGauge(descNodeStatusPhase, boolFloat64(p == v1.NodeTerminated), string(v1.NodeTerminated))
	}

	if !nc.opts.DisableNodeNonGenericResourceMetrics {
		// Add capacity and allocatable resources if they are set.
		addResource := func(d *prometheus.Desc, res v1.ResourceList, n v1.ResourceName) {
			if v, ok := res[n]; ok {
				addGauge(d, float64(v.MilliValue())/1000)
			}
		}

		addResource(descNodeStatusCapacityCPU, n.Status.Capacity, v1.ResourceCPU)
		addResource(descNodeStatusCapacityMemory, n.Status.Capacity, v1.ResourceMemory)
		addResource(descNodeStatusCapacityPods, n.Status.Capacity, v1.ResourcePods)

		addResource(descNodeStatusAllocatableCPU, n.Status.Allocatable, v1.ResourceCPU)
		addResource(descNodeStatusAllocatableMemory, n.Status.Allocatable, v1.ResourceMemory)
		addResource(descNodeStatusAllocatablePods, n.Status.Allocatable, v1.ResourcePods)
	}

	capacity := n.Status.Capacity
	allocatable := n.Status.Allocatable

	for resourceName, val := range capacity {
		switch resourceName {
		case v1.ResourceCPU:
			addGauge(descNodeStatusCapacity, float64(val.MilliValue())/1000, sanitizeLabelName(string(resourceName)), string(constant.UnitCore))
		case v1.ResourceStorage:
			fallthrough
		case v1.ResourceEphemeralStorage:
			fallthrough
		case v1.ResourceMemory:
			addGauge(descNodeStatusCapacity, float64(val.MilliValue())/1000, sanitizeLabelName(string(resourceName)), string(constant.UnitByte))
		case v1.ResourcePods:
			addGauge(descNodeStatusCapacity, float64(val.MilliValue())/1000, sanitizeLabelName(string(resourceName)), string(constant.UnitInteger))
		default:
			if helper.IsHugePageResourceName(resourceName) {
				addGauge(descNodeStatusCapacity, float64(val.MilliValue())/1000, sanitizeLabelName(string(resourceName)), string(constant.UnitByte))
			}
			if helper.IsAttachableVolumeResourceName(resourceName) {
				addGauge(descNodeStatusCapacity, float64(val.MilliValue())/1000, sanitizeLabelName(string(resourceName)), string(constant.UnitByte))
			}
			if helper.IsExtendedResourceName(resourceName) {
				addGauge(descNodeStatusCapacity, float64(val.MilliValue())/1000, sanitizeLabelName(string(resourceName)), string(constant.UnitInteger))
			}
		}
	}

	for resourceName, val := range allocatable {
		switch resourceName {
		case v1.ResourceCPU:
			addGauge(descNodeStatusAllocatable, float64(val.MilliValue())/1000, sanitizeLabelName(string(resourceName)), string(constant.UnitCore))
		case v1.ResourceStorage:
			fallthrough
		case v1.ResourceEphemeralStorage:
			fallthrough
		case v1.ResourceMemory:
			addGauge(descNodeStatusAllocatable, float64(val.MilliValue())/1000, sanitizeLabelName(string(resourceName)), string(constant.UnitByte))
		case v1.ResourcePods:
			addGauge(descNodeStatusAllocatable, float64(val.MilliValue())/1000, sanitizeLabelName(string(resourceName)), string(constant.UnitInteger))
		default:
			if helper.IsHugePageResourceName(resourceName) {
				addGauge(descNodeStatusAllocatable, float64(val.MilliValue())/1000, sanitizeLabelName(string(resourceName)), string(constant.UnitByte))
			}
			if helper.IsAttachableVolumeResourceName(resourceName) {
				addGauge(descNodeStatusAllocatable, float64(val.MilliValue())/1000, sanitizeLabelName(string(resourceName)), string(constant.UnitByte))
			}
			if helper.IsExtendedResourceName(resourceName) {
				addGauge(descNodeStatusAllocatable, float64(val.MilliValue())/1000, sanitizeLabelName(string(resourceName)), string(constant.UnitInteger))
			}
		}
	}

}
