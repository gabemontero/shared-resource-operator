/*
Copyright 2023.

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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	sharedresourcesdeveloperredhatcomv1 "github.com/gabemontero/shared-resource-operator/api/v1"
	mfc "github.com/manifestival/controller-runtime-client"
	"github.com/manifestival/manifestival"
)

// ConfigReconciler reconciles a Config object
type ConfigReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	Manifest    manifestival.Manifest
	manifestErr bool
}

//+kubebuilder:rbac:groups=sharedresources.developer.redhat.com.developer.redhat.com,resources=configs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=sharedresources.developer.redhat.com.developer.redhat.com,resources=configs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=sharedresources.developer.redhat.com.developer.redhat.com,resources=configs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Config object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *ConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if r.manifestErr {
		logger.Info("Retrying applying manifest's resources...")
		r.manifestErr = false
		if err := r.Manifest.Apply(); err != nil {
			logger.Error(err, "rolling out manifest's resources")
			return ctrl.Result{Requeue: true}, err
		}
		logger.Info("Worked this time")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := r.setupManifestival(mgr.GetLogger()); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&sharedresourcesdeveloperredhatcomv1.Config{}).
		Complete(r)
}

func (r *ConfigReconciler) setupManifestival(logger logr.Logger) error {
	logger.Info("GGM in setupManifestival")
	client := mfc.NewClient(r.Client)

	var err error
	r.Manifest, err = manifestival.NewManifest(
		"/tmp/assets/release.yaml",
		manifestival.UseClient(client),
		manifestival.UseLogger(logger),
	)
	if err != nil {
		logger.Error(err, "got NewManifest error: %s", err.Error())
		return err
	}
	if err = r.Manifest.Apply(); err != nil {
		logger.Error(err, "problem applying release.yaml")
		r.manifestErr = true
	} else {
		logger.Info("manifest apply was OK")
	}
	// we will retry in reconciler
	return nil

}
