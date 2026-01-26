// Package kube provides the kubernetes integration for private locations
package kube

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v3 "github.com/Escape-Technologies/cli/pkg/api/v3"
	"github.com/Escape-Technologies/cli/pkg/log"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/proxy"
)

const (
	defaultPort         = 8001
	defaultStaticPrefix = "/static/"
	defaultAPIPrefix    = "/"
	defaultAddress      = "127.0.0.1"
)

func inferConfig() (*rest.Config, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	var c *rest.Config
	var err error

	if kubeconfig != "" {
		log.Trace("Using kubeconfig : %s", kubeconfig)
		c, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build config from kubeconfig: %w", err)
		}
	} else {
		log.Trace("Using in cluster config")
		c, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to build in cluster config: %w", err)
		}
	}
	return c, nil
}

func connectAndRun(ctx context.Context, cfg *rest.Config, isConnected *atomic.Bool, locationID string, locationName string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	srv, err := proxy.NewServer(
		"",
		"/",
		defaultStaticPrefix,
		nil,
		cfg,
		0,
		false,
	)
	if err != nil {
		return fmt.Errorf("error creating proxy server: %w", err)
	}

	lis, err := srv.Listen(defaultAddress, defaultPort)
	if err != nil {
		return fmt.Errorf("error listening: %w", err)
	}

	go func() {
		for !isConnected.Load() || ctx.Err() != nil {
			time.Sleep(1 * time.Second)
		}
		if ctx.Err() != nil {
			lis.Close() //nolint:errcheck
			return
		}
		log.Info("Connected to K8s API")
		log.Trace("Upserting K8s integration")
		req := v3.NewCreatekubernetesIntegrationRequest(
			locationName,
			*v3.NewCreatekubernetesIntegrationRequestParameters(),
		)
		req.ProxyId = &locationID
		_, err := escape.UpsertKubernetesIntegration(ctx, *req)
		if err != nil {
			errMsg := fmt.Sprintf("%s", err)
			log.Error("Failed to register Kubernetes integration: %s", errMsg)
			if strings.Contains(errMsg, "401") || strings.Contains(errMsg, "unauthorized") {
				log.Error("Check your ESCAPE_API_KEY environment variable")
				log.Error("Get your API key from https://app.escape.tech/organization/settings")
			}
			return
		}

		<-ctx.Done()
		lis.Close() //nolint:errcheck
	}()

	log.Debug("Connecting to k8s API")
	err = srv.ServeOnListener(lis)
	if err != nil {
		return fmt.Errorf("error serving: %w", err)
	}
	return nil
}

// Start the kubernetes integration
func Start(ctx context.Context, locationID string, locationName string, healthy *atomic.Bool) {
	cfg, err := inferConfig()
	if err != nil {
		log.Debug("Error inferring kubeconfig: %s", err)
		log.Info("Not connected to k8s API")
		return
	}
	for {
		err = connectAndRun(ctx, cfg, healthy, locationID, locationName)
		if err != nil {
			log.Error("Failed to connect to Kubernetes API: %s", err)
		}
		if ctx.Err() != nil {
			return
		}
	}
}
