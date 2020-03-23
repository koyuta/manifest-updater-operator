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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	manifestupdaterkoyutaiov1alpha1 "manifest-updater/api/v1alpha1"
	"manifest-updater/updater"
)

const (
	defaultBranch = "master"
)

const (
	imageTagRegexp = `( *)(?P<tag>\w[\w-\.]{0,127})`
)

// UpdaterReconciler reconciles a Updater object
type UpdaterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Queue  chan<- *updater.Entry
}

// +kubebuilder:rbac:groups=manifest-updater.koyuta.io.koyuta.io,resources=updaters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=manifest-updater.koyuta.io.koyuta.io,resources=updaters/status,verbs=get;update;patch

func (r *UpdaterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ := r.Log.WithValues("updater", req.NamespacedName)

	u := &manifestupdaterkoyutaiov1alpha1.Updater{}
	if err := r.Get(ctx, req.NamespacedName, u); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
	}

	setDefaultValuesIfNotPresent(u)

	entry := &updater.Entry{
		DockerHub: u.Spec.Registry.DockerHub,
		Filter:    u.Spec.Registry.Filter,
		Git:       u.Spec.Repository.Git,
		Branch:    u.Spec.Repository.Branch,
		Path:      u.Spec.Repository.Path,
	}
	r.Queue <- entry

	return ctrl.Result{}, nil
}

func setDefaultValuesIfNotPresent(updater *manifestupdaterkoyutaiov1alpha1.Updater) {
	if updater.Spec.Repository.Branch == "" {
		updater.Spec.Repository.Branch = defaultBranch
	}
}

func (r *UpdaterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&manifestupdaterkoyutaiov1alpha1.Updater{}).
		Complete(r)
}
