// Copyright (c) 2020 Red Hat, Inc.

package webhook

import (
	"github.com/prometheus/common/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// Setup - adds the webhook server to manager
func Setup(mgr manager.Manager) error {
	hookServer := &webhook.Server{
		Port:    6443,
		CertDir: "/tmp/webhookcert",
	}

	log.Info("Add the webhook server.")
	if err := mgr.Add(hookServer); err != nil {
		return err
	}

	log.Info("Registering webhooks to the webhook server.")
	validatingPath := "/validate-v1-klusterletaddonconfig"
	hookServer.Register(validatingPath, &webhook.Admission{Handler: &klusterletAddonConfigValidator{}})

	return nil
}
