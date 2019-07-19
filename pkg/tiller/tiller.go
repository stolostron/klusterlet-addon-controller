package tiller

import (
	"context"

	klusterletv1alpha1 "github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/apis/klusterlet/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("tiller")

func Reconcile(instance *klusterletv1alpha1.KlusterletService, client client.Client, scheme *runtime.Scheme) error {
	tiller := newTillerCR(instance)
	err := controllerutil.SetControllerReference(instance, tiller, scheme)
	if err != nil {
		return err
	}

	foundTiller := &klusterletv1alpha1.Tiller{}
	err = client.Get(context.TODO(), types.NamespacedName{Name: tiller.Name, Namespace: tiller.Namespace}, foundTiller)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Tiller", "Tiller.Namespace", tiller.Namespace, "Tiller.Name", tiller.Name)
		err = client.Create(context.TODO(), tiller)
		if err != nil {
			return err
		}
	}
	return nil
}

func newTillerCR(cr *klusterletv1alpha1.KlusterletService) *klusterletv1alpha1.Tiller {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &klusterletv1alpha1.Tiller{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-tiller",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: klusterletv1alpha1.TillerSpec{
			FullNameOverride: cr.Name + "-tiller",
			CACertIssuer:     cr.Name + "-self-signed",
			DefaultAdminUser: cr.Name + "admin",
		},
	}
}
