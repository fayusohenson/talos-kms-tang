package server

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/anatol/clevis.go"
	"github.com/siderolabs/kms-client/api/kms"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	kms.UnimplementedKMSServiceServer

	tangConfig *TangConfig
	logger     *slog.Logger
}

func NewServer(tc *TangConfig, logger *slog.Logger) *Server {
	return &Server{
		tangConfig: tc,
		logger:     logger,
	}
}

func (s *Server) Seal(ctx context.Context, req *kms.Request) (*kms.Response, error) {
	if req.Data == nil {
		return nil, status.Errorf(codes.InvalidArgument, "data is required")
	}
	if req.NodeUuid == "" {
		return nil, status.Errorf(codes.InvalidArgument, "node_uuid is required")
	}

	s.logger.Info("Sealing data", "node", req.NodeUuid)

	_, err := s.tangConfig.IsValidThumbprint(ctx)
	if err != nil {
		s.logger.Error("Error while verifying key", "error", err)
		return nil, status.Errorf(codes.Internal, "Internal Error")
	}

	tangConfigBytes, err := json.Marshal(s.tangConfig)
	if err != nil {
		s.logger.Error("Error while marshaling tang config", "error", err)
		return nil, status.Errorf(codes.Internal, "Internal Error")
	}

	dataToEncryptBytes, err := json.Marshal(DataPayload{
		Data:     req.Data,
		NodeUuid: req.NodeUuid,
	})
	if err != nil {
		s.logger.Error("Error while marshaling data before sealing", "error", err)
		return nil, status.Errorf(codes.Internal, "Internal Error")
	}

	encrypted, err := clevis.Encrypt(dataToEncryptBytes, "tang", string(tangConfigBytes))
	if err != nil {
		s.logger.Error("Error while sealing data", "error", err)
		return nil, status.Errorf(codes.Internal, "Internal Error")
	}

	return &kms.Response{
		Data: encrypted,
	}, nil
}

func (s *Server) Unseal(ctx context.Context, req *kms.Request) (*kms.Response, error) {
	if req.Data == nil {
		return nil, status.Errorf(codes.InvalidArgument, "data is required")
	}
	if req.NodeUuid == "" {
		return nil, status.Errorf(codes.InvalidArgument, "node_uuid is required")
	}

	s.logger.Info("Unsealing data", "node", req.NodeUuid)

	decrypted, err := clevis.Decrypt(req.Data)
	if err != nil {
		s.logger.Error("Error while unsealing data", "error", err)
		return nil, status.Errorf(codes.Internal, "Internal Error")
	}

	var originalData DataPayload
	if err := json.Unmarshal(decrypted, &originalData); err != nil {
		s.logger.Error("Error while unmarshaling data", "error", err)
		return nil, status.Errorf(codes.Internal, "Internal Error")
	}

	if originalData.NodeUuid != req.NodeUuid {
		s.logger.Error(
			"Node UUID does not match",
			"request_node",
			req.NodeUuid,
			"original_node",
			originalData.NodeUuid,
		)
		return nil, status.Errorf(codes.PermissionDenied, "Forbidden")
	}

	return &kms.Response{
		Data: originalData.Data,
	}, nil
}
