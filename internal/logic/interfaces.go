package logic

import (
	"github.com/res-am/grpc-fts/internal/config"
	"github.com/res-am/grpc-fts/internal/models"
	"google.golang.org/grpc/status"
)

type Runner interface {
	RunTestCases() error
}

type ResponseChecker interface {
	CheckResponse(response map[string]interface{}, expectations map[string]interface{}) ([]models.ValidationFail, error)
	CheckStatus(*status.Status, *config.Status) ([]models.ValidationFail, error)
}
