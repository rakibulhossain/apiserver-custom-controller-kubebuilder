package controllers

import (
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	batchv1 "github.com/Rakibul-Hossain/api/v1"
)

func newDeployment(custom *batchv1.Custom) *appsv1.Deployment {
	labels := map[string]string{
		"app":        "main",
		"controller": custom.Name,
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      custom.Spec.DeploymentName,
			Namespace: custom.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(custom, batchv1.GroupVersion.WithKind("Custom")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: custom.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "hello",
							Image: custom.Spec.Image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: *custom.Spec.ContainerPort,
								},
							},
						},
					},
				},
			},
		},
	}
}

func newService(custom *batchv1.Custom) *corev1.Service {
	fmt.Println("newService is called")
	labels := map[string]string{
		"app":        "main",
		"controller": custom.Name,
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      custom.Spec.DeploymentName + "-service",
			Namespace: custom.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(custom, batchv1.GroupVersion.WithKind("Custom")),
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Type:     "NodePort",
			Ports: []corev1.ServicePort{
				{
					NodePort: int32(30011),
					Port:     *custom.Spec.ServicePort,
					TargetPort: intstr.IntOrString{
						IntVal: *custom.Spec.TargetPort,
					},
				},
			},
		},
	}
}
