package main

import (
	"context"

	"github.com/networkservicemesh/networkservicemesh/controlplane/api/connection/mechanisms/kernel"
	"github.com/networkservicemesh/networkservicemesh/pkg/tools"
	"github.com/networkservicemesh/networkservicemesh/pkg/tools/spanhelper"

	"github.com/sirupsen/logrus"

	"github.com/networkservicemesh/networkservicemesh/pkg/tools/jaeger"

	"github.com/networkservicemesh/networkservicemesh/sdk/common"

	"github.com/networkservicemesh/networkservicemesh/sdk/client"
)

type nsmClientApp struct {
	configuration *common.NSConfiguration
}

func (c *nsmClientApp) Run() {
	closer := jaeger.InitJaeger("nsm-init")
	defer func() { _ = closer.Close() }()

	span := spanhelper.FromContext(context.Background(), "RequestNetworkService")
	defer span.Finish()

	c.configuration = c.configuration.FromEnv()
	if c.configuration.PodName == "" {
		podName, err := tools.GetCurrentPodNameFromHostname()
		if err != nil {
			logrus.Infof("failed to get current pod name from hostname: %v", err)
		} else {
			c.configuration.PodName = podName
		}
	}
	if c.configuration.Namespace == "" {
		c.configuration.Namespace = common.GetNamespace()
	}

	clientList, err := client.NewNSMClientList(span.Context(), c.configuration)
	if err != nil {
		span.Finish()
		_ = closer.Close()
		logrus.Fatalf("nsm client: Unable to create the NSM client %v", err)
		return
	}
	err = clientList.ConnectRetry(span.Context(), "nsm", kernel.MECHANISM, "Primary interface", client.ConnectionRetry, client.RequestDelay)
	if err != nil {
		span.Finish()
		_ = closer.Close()
		logrus.Fatalf("nsm client: Unable to establish connection with network service")
		return
	}
	logrus.Info("nsm client: initialization is completed successfully")
	
	// Capture signals to cleanup before exiting
	ch := tools.NewOSSignalChannel()
	logrus.Info("nsm client: all clients connected. Sleeping...")
	<-ch
}

// NewNSMClientApp - creates a client application.
func NewNSMClientApp(configration *common.NSConfiguration) *nsmClientApp {
	return &nsmClientApp{
		configuration: configration,
	}
}
