package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"

	"github.com/fayusohenson/talos-kms-tang/pkg/server"
	"github.com/siderolabs/kms-client/api/kms"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

var kmsFlags struct {
	apiEndpoint    string
	tlsCertPath    string
	tlsKeyPath     string
	tlsEnable      bool
	tangEndpoint   string
	tangThumbprint string
}

func main() {
	flag.StringVar(&kmsFlags.apiEndpoint, "kms-api-endpoint", ":4050", "gRPC API endpoint for the KMS")
	flag.BoolVar(&kmsFlags.tlsEnable, "tls-enable", false, "whether to enable tls or not")
	flag.StringVar(&kmsFlags.tlsCertPath, "tls-cert-path", "", "path to TLS certificate file")
	flag.StringVar(&kmsFlags.tlsKeyPath, "tls-key-path", "", "path to TLS private key file")
	flag.StringVar(&kmsFlags.tangEndpoint, "tang-endpoint", "", "tang server endpoint")
	flag.StringVar(&kmsFlags.tangThumbprint, "tang-thumbprint", "", "thumbprint of a trusted signing key")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx, logger); err != nil {
		logger.Error("Error during initialization", "error", err)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	if kmsFlags.tangEndpoint == "" {
		return fmt.Errorf("--tang-endpoint is not set")
	}

	logger.Info("Initializing...")

	srv := server.NewServer(&server.TangConfig{
		URL:        kmsFlags.tangEndpoint,
		Thumbprint: kmsFlags.tangThumbprint,
	}, logger)

	var s *grpc.Server

	if kmsFlags.tlsEnable {
		var creds credentials.TransportCredentials

		creds, err := credentials.NewServerTLSFromFile(kmsFlags.tlsCertPath, kmsFlags.tlsKeyPath)
		if err != nil {
			return fmt.Errorf("failed to create TLS credentials: %w", err)
		}

		logger.Info("TLS enabled")

		s = grpc.NewServer(grpc.Creds(creds))
	} else {
		s = grpc.NewServer()
	}

	kms.RegisterKMSServiceServer(s, srv)
	reflection.Register(s)

	lis, err := net.Listen("tcp", kmsFlags.apiEndpoint)
	if err != nil {
		return fmt.Errorf("error listening for gRPC API: %w", err)
	}

	logger.Info("Starting server...", "endpoint", kmsFlags.apiEndpoint)

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return s.Serve(lis)
	})

	eg.Go(func() error {
		<-ctx.Done()

		logger.Info("Stopping server...")

		s.Stop()

		return nil
	})

	if err := eg.Wait(); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return err
	}

	return nil
}
