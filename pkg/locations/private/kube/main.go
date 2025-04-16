package kube

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/Escape-Technologies/cli/pkg/api/escape"
	v2 "github.com/Escape-Technologies/cli/pkg/api/v2"
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

	if kubeconfig != "" {
		log.Trace("Using kubeconfig : %s", kubeconfig)
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		log.Trace("Using in cluster config")
		return rest.InClusterConfig()
	}
}

func connectAndRun(ctx context.Context, cfg *rest.Config, isConnected *atomic.Bool, locationId string, locationName string) error {
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
			lis.Close()
			return
		}
		log.Info("Connected to k8s API")
		log.Trace("Upserting k8s integration")
		err = escape.UpsertIntegration(ctx, &v2.UpdateIntegrationRequest{
			Name:       locationName,
			LocationId: &locationId,
		})
		if err != nil {
			log.Error("Error upserting integration: %s", err)
			return
		}

		<-ctx.Done()
		lis.Close()
	}()

	log.Debug("Connecting to k8s API")
	err = srv.ServeOnListener(lis)
	if err != nil {
		return fmt.Errorf("error serving: %w", err)
	}
	return nil
}

func Start(ctx context.Context, locationId string, locationName string, healthy *atomic.Bool) {
	cfg, err := inferConfig()
	if err != nil {
		log.Debug("Error inferring kubeconfig: %s", err)
		log.Info("Not connected to k8s API")
		return
	}
	for {
		err = connectAndRun(ctx, cfg, healthy, locationId, locationName)
		if err != nil {
			log.Error("Error connecting to k8s API: %s", err)
		}
		if ctx.Err() != nil {
			return
		}
	}

}
