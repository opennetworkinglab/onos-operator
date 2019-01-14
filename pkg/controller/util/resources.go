package util

import (
	"fmt"
	"github.com/opennetworkinglab/onos-operator/pkg/apis/onos/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
	"strconv"
)

const (
	AppKey     = "app"
	ClusterKey = "cluster"
	OnosApp    = "onos"
)

const (
	InitScriptsVolume  = "init-scripts"
	ProbeScriptsVolume = "probe-scripts"
	ConfigVolume       = "config"
)

const (
	AtomixSuffix = "atomix"
	InitSuffix   = "init"
	ProbeSuffix  = "probe"
)

// onosPodRegex is a regular expression that extracts the parent StatefulSet and ordinal from the Name of a Pod
var onosPodRegex = regexp.MustCompile("(.*)-([0-9]+)$")

// getParentNameAndOrdinal gets the name of pod's parent StatefulSet and pod's ordinal as extracted from its Name. If
// the Pod was not created by a StatefulSet, its parent is considered to be empty string, and its ordinal is considered
// to be -1.
func getParentNameAndOrdinal(pod *corev1.Pod) (string, int) {
	parent := ""
	ordinal := -1
	subMatches := onosPodRegex.FindStringSubmatch(pod.Name)
	if len(subMatches) < 3 {
		return parent, ordinal
	}
	parent = subMatches[1]
	if i, err := strconv.ParseInt(subMatches[2], 10, 32); err == nil {
		ordinal = int(i)
	}
	return parent, ordinal
}

// getParentName gets the name of pod's parent StatefulSet. If pod has not parent, the empty string is returned.
func getParentName(pod *corev1.Pod) string {
	parent, _ := getParentNameAndOrdinal(pod)
	return parent
}

//  getOrdinal gets pod's ordinal. If pod has no ordinal, -1 is returned.
func getOrdinal(pod *corev1.Pod) int {
	_, ordinal := getParentNameAndOrdinal(pod)
	return ordinal
}

// getPodName gets the name of set's child Pod with an ordinal index of ordinal
func getPodName(set *v1alpha1.OnosCluster, ordinal int) string {
	return fmt.Sprintf("%s-%d", set.Name, ordinal)
}

// getClusterLabels returns the labels for an ONOS cluster pod
func getClusterLabels(cluster *v1alpha1.OnosCluster) map[string]string {
	return map[string]string{
		AppKey:     OnosApp,
		ClusterKey: cluster.Name,
	}
}

func getResourceName(cluster *v1alpha1.OnosCluster, resource string) string {
	return fmt.Sprintf("%s-%s", cluster.Name, resource)
}

func getOnosServiceName(cluster *v1alpha1.OnosCluster) string {
	return cluster.Name
}

func getInitConfigMapName(cluster *v1alpha1.OnosCluster) string {
	return getResourceName(cluster, InitSuffix)
}

func getAtomixServiceName(cluster *v1alpha1.OnosCluster) string {
	return getResourceName(cluster, AtomixSuffix)
}

func getInitScriptsVolumeName(cluster *v1alpha1.OnosCluster) string {
	return getResourceName(cluster, InitSuffix)
}

func getProbeScriptsVolumeName(cluster *v1alpha1.OnosCluster) string {
	return getResourceName(cluster, ProbeSuffix)
}

// NewOnosService returns a new headless service for the ONOS cluster
func NewOnosService(cluster *v1alpha1.OnosCluster) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getOnosServiceName(cluster),
			Namespace: cluster.Namespace,
			Labels:    getClusterLabels(cluster),
			Annotations: map[string]string{
				"service.alpha.kubernetes.io/tolerate-unready-endpoints": "true",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: cluster.Name + "-node",
					Port: 5679,
				},
			},
			PublishNotReadyAddresses: true,
			ClusterIP:                "None",
			Selector: map[string]string{
				AppKey: cluster.Name,
			},
		},
	}
}

// NewInitConfigMap returns a new ConfigMap for initializing ONOS nodes
func NewInitConfigMap(cluster *v1alpha1.OnosCluster) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getInitConfigMapName(cluster),
			Namespace: cluster.Namespace,
			Labels:    getClusterLabels(cluster),
		},
		Data: map[string]string{
			"create_config.sh": getInitConfigMapScript(cluster),
		},
	}
}

// getInitConfigMapScript returns a new script for generating an Atomix configuration
func getInitConfigMapScript(cluster *v1alpha1.OnosCluster) string {
	return `
#!/usr/bin/env bash
apt-get update > /dev/null 2>&1
apt-get -y install dnsutils > /dev/null 2>&1
USER=$(whoami)
HOST=$(hostname -s)
DOMAIN=$(hostname -d)
CONFIG_DIR="/root/onos/config"
CONFIG_FILE="$CONFIG_DIR/cluster.json"
HEAP=2G
function print_usage() {
    echo "\
    Usage: start-onos [OPTIONS]
    Starts an ONOS node based on the supplied options.
    --service           The name of the Atomix service to which to connect.
    "
}
function print_config() {
    echo "{"
    print_node
    print_storage
    echo "}"
}
function print_node() {
    echo "  \"node\": {"
    echo "    \"id\": \"$HOST\","
    echo "    \"host\": \"$HOST\","
    echo "    \"port\": 9876"
    echo "  },"
}
function print_storage() {
    echo "  \"service\": \"$ATOMIX_CLUSTER-service\","
}
ATOMIX_CLUSTER=$1
until nslookup "$ATOMIX_CLUSTER-service" > /dev/null 2>&1; do sleep 2; done;
print_config`
}

// NewProbeConfigMap returns a new ConfigMap for probing ONOS nodes
func NewProbeConfigMap(cluster *v1alpha1.OnosCluster) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getInitConfigMapName(cluster),
			Namespace: cluster.Namespace,
			Labels:    getClusterLabels(cluster),
		},
		Data: map[string]string{
			"check-onos-status": getProbeConfigMapScript(cluster),
		},
	}
}

// getProbeConfigMapScript returns a script string for probing ONOS nodes
func getProbeConfigMapScript(cluster *v1alpha1.OnosCluster) string {
	return `
#!/bin/bash
set -e
host=$(hostname -s)
config=$(curl -s http://$host:8181/onos/v1/cluster/$host --user onos:rocks)
echo $config
printf '%q' $config | grep -q "READY"`
}

// NewOnosPod returns a new pod for an ONOS node
func NewOnosPod(cluster *v1alpha1.OnosCluster, ordinal int) *corev1.Pod {
	atomixCluster := cluster.Spec.Atomix.Service
	if atomixCluster == "" {
		atomixCluster = getAtomixServiceName(cluster)
	}
	defaultMode := int32(0744)
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getPodName(cluster, ordinal),
			Namespace: cluster.Namespace,
			Labels:    getClusterLabels(cluster),
		},
		Spec: corev1.PodSpec{
			InitContainers: newInitContainers(atomixCluster),
			Containers:     newContainers(cluster.Spec.Env, cluster.Spec.Resources, cluster.Spec.Apps),
			Volumes: []corev1.Volume{
				{
					Name: InitScriptsVolume,
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: getInitScriptsVolumeName(cluster),
							},
							DefaultMode: &defaultMode,
						},
					},
				},
				{
					Name: ProbeScriptsVolume,
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: getProbeScriptsVolumeName(cluster),
							},
							DefaultMode: &defaultMode,
						},
					},
				},
				{
					Name: ConfigVolume,
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
			},
		},
	}
}

func newInitContainers(atomixCluster string) []corev1.Container {
	return []corev1.Container{
		newInitContainer(atomixCluster),
	}
}

func newInitContainer(atomixCluster string) corev1.Container {
	return corev1.Container{
		Name:  "configure",
		Image: "ubuntu:16.04",
		Env: []corev1.EnvVar{
			{
				Name:  "ATOMIX_CLUSTER",
				Value: atomixCluster,
			},
		},
		Command: []string{
			"bash",
			"-c",
			"/scripts/configure-onos.sh $ATOMIX_CLUSTER > /config/cluster.json && touch /config/active",
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      InitScriptsVolume,
				MountPath: "/scripts",
			},
			{
				Name:      ConfigVolume,
				MountPath: "/config",
			},
		},
	}
}

func newContainers(env []corev1.EnvVar, resources corev1.ResourceRequirements, apps []string) []corev1.Container {
	return []corev1.Container{
		newContainer(env, resources, apps),
	}
}

func newContainer(env []corev1.EnvVar, resources corev1.ResourceRequirements, apps []string) corev1.Container {
	privileged := true
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      ProbeScriptsVolume,
			MountPath: "/root/onos/bin/check-onos-status",
			SubPath: "check-onos-status",
		},
		{
			Name:      ConfigVolume,
			MountPath: "/root/onos/config/cluster.json",
			SubPath:   "cluster.json",
		},
	}

	for _, app := range apps {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      ConfigVolume,
			MountPath: "/root/onos/apps/" + app + "/active",
			SubPath:   "active",
		})
	}

	return corev1.Container{
		Name:            "onos",
		Image:           "onosproject/onos:2.0.0-rc2",
		ImagePullPolicy: corev1.PullAlways,
		Env:             env,
		Resources:       resources,
		Ports: []corev1.ContainerPort{
			{
				Name:          "openflow",
				ContainerPort: 6653,
			},
			{
				Name:          "ovsdb",
				ContainerPort: 6640,
			},
			{
				Name:          "east-west",
				ContainerPort: 9876,
			},
			{
				Name:          "cli",
				ContainerPort: 8101,
			},
			{
				Name:          "ui",
				ContainerPort: 8181,
			},
		},
		SecurityContext: &corev1.SecurityContext{
			Privileged: &privileged,
		},
		ReadinessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{
						"sh",
						"-c",
						"/root/onos/bin/check-onos-status",
					},
				},
			},
			InitialDelaySeconds: 30,
			TimeoutSeconds:      15,
			FailureThreshold:    10,
		},
		LivenessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{
						"sh",
						"-c",
						"/root/onos/bin/check-onos-status",
					},
				},
			},
			InitialDelaySeconds: 300,
			TimeoutSeconds:      15,
			FailureThreshold:    5,
		},
		VolumeMounts: volumeMounts,
	}
}
