package kube

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/Escape-Technologies/cli/pkg/log"
	"github.com/oapi-codegen/runtime/types"
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
		log.Debug("Using kubeconfig : %s", kubeconfig)
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		log.Debug("Using in cluster config")
		return rest.InClusterConfig()
	}
}

func connectAndRun(ctx context.Context, cfg *rest.Config, isConnected *atomic.Bool) error {
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
		<-ctx.Done()
		lis.Close()
	}()

	log.Debug("Connecting to k8s API")
	isConnected.Store(true)
	err = srv.ServeOnListener(lis)
	if err != nil {
		return fmt.Errorf("error serving: %w", err)
	}
	return nil
}


func Start(ctx context.Context, locationId *types.UUID, locationName string, healthy *atomic.Bool) {
	cfg, err := inferConfig()
	if err != nil {
		log.Info("Not connected to k8s API")
		return
	}
	for {
		err = connectAndRun(ctx, cfg, healthy)
		if err != nil {
			log.Error("Error connecting to k8s API: %s", err)
			return
		}
		time.Sleep(1 * time.Second)
		if ctx.Err() != nil {
			return
		}
		if healthy.Load() {
			err = UpsertIntegration(ctx, locationId, locationName)
			if err != nil {
				log.Error("Error upserting integration: %s", err)
				return
			}
		}
	}

	
}

