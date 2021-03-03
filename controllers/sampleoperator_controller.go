/*


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

package controllers

import (
	"context"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	sampleoperatorv1 "sample-operator/api/v1"
)

const (
	UsernameKey = "user"
	PasswordKey = "password"
)


// SampleOperatorReconciler reconciles a SampleOperator object
type SampleOperatorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=sample-operator.example.com,resources=sampleoperators,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=sample-operator.example.com,resources=sampleoperators/status,verbs=get;update;patch

func (r *SampleOperatorReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("sampleoperator", req.NamespacedName)

	data := map[string]string{}
	data[UsernameKey] = "sample-user"
	data[PasswordKey] = "sample-pwd"

	// Fetch SampleOperator Instance
	instance := &sampleoperatorv1.SampleOperator{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		// IsNotFound returns true if the specified error was created by NewNotFound.
		// Request object not found, could have been deleted after reconcile request.
		// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
		// Return and don't requeue
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	err = r.ensureLatestPod(instance, data)
	if err != nil {
		return reconcile.Result{}, err
	}

	configMap := r.newConfigMapForCr(instance, data)
	// Set Database instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, configMap, r.Scheme); err != nil {
		return reconcile.Result{}, err
	}
	// Check if this ConfigMap already exists
	configMapFound := &corev1.ConfigMap{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: configMap.Name, Namespace: configMap.Namespace}, configMapFound)
	if err != nil && errors.IsNotFound(err) {
		r.Log.Info("Creating a new ConfigMap", "ConfigMap.Namespace", configMap.Namespace, "ConfigMap.Name", configMap.Name)
		err = r.Client.Create(context.TODO(), configMap)
		if err != nil {
			return reconcile.Result{}, err
		}
		configMapFound = configMap
	} else if err != nil {
		return reconcile.Result{}, err
	}
	// ConfigMap created successfully - update status with the reference
	instance.Status.SampleConfigMap = configMapFound.Name
	// Update status
	err = r.Client.Status().Update(context.TODO(), instance)
	if err != nil {
		r.Log.Error(err, "Failed to update status with DBConfigMap")
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *SampleOperatorReconciler) ensureLatestConfigMap(instance *sampleoperatorv1.SampleOperator, data map[string]string) (bool, error) {
	configMap := r.newConfigMapForCr(instance, data)
	// Set SampleOperator instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, configMap, r.Scheme); err != nil {
		return false, err
	}
	// Check if this ConfigMap already exists
	foundMap := &corev1.ConfigMap{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: configMap.Name, Namespace: configMap.Namespace}, foundMap)
	if err != nil && errors.IsNotFound(err) {
		err = r.Client.Create(context.TODO(), configMap)
		if err != nil {
			return false, err
		}
	} else if err != nil {
		return false, err
	}

	if foundMap.Data["user"] != configMap.Data["user"] {
		err = r.Client.Update(context.TODO(), configMap)
		if err != nil {
			return false, err
		}
		// ConfigMap created successfully - update status with the reference
		instance.Status.SampleConfigMap = foundMap.Name
		// Update status
		err = r.Client.Status().Update(context.TODO(), instance)
		if err != nil {
			r.Log.Error(err, "Failed to update status with DBConfigMap")
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (r *SampleOperatorReconciler) newConfigMapForCr(instance *sampleoperatorv1.SampleOperator, data map[string]string) *corev1.ConfigMap {
	labels := map[string]string{
		"app": instance.Name,
	}
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "config",
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Data: data,
	}
	return configMap
}

func (r *SampleOperatorReconciler) ensureLatestPod(instance *sampleoperatorv1.SampleOperator, data map[string]string) error {
	// Define a new Pod object
	pod := r.newPodForCR(instance, data)

	// Set Presentation instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, pod, r.Scheme); err != nil {
		return err
	}
	// Check if this Pod already exists
	found := &corev1.Pod{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		err = r.Client.Create(context.TODO(), pod)
		if err != nil {
			return err
		}
		// Pod created successfully - don't requeue
		return nil
	} else if err != nil {

		return err
	}
	return nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func (r *SampleOperatorReconciler) newPodForCR(instance *sampleoperatorv1.SampleOperator, data map[string]string) *corev1.Pod {
	labels := map[string]string{
		"app": instance.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-pod",
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "slides",
					Image: "manueldewald/presentation",
					Env: []corev1.EnvVar{
						{
							Name:  "SAMPLE_SERVICE_USER",
							Value: data[UsernameKey],
						},
						{
							Name:  "SAMPLE_SERVICE_PASSWORD",
							Value: data[PasswordKey],
						},
					},
				},
			},
		},
	}
}


func (r *SampleOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sampleoperatorv1.SampleOperator{}).
		Complete(r)
}
