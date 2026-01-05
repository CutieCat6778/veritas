package utils

import (
	"errors"
	"context"

	"gorm.io/gorm"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"github.com/99designs/gqlgen/graphql"
)

// HandleGormError processes GORM errors and returns a user-friendly message and HTTP status code
func HandleGormError(err error) (string, int) {
	if err == nil {
		return "", 0
	}

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return "Record not found", 404
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return "Duplicate key error: resource already exists", 409
	case errors.Is(err, gorm.ErrForeignKeyViolated):
		return "Foreign key violation: invalid reference", 400
	case errors.Is(err, gorm.ErrInvalidData):
		return "Invalid data provided", 400
	case errors.Is(err, gorm.ErrNotImplemented):
		return "Operation not supported", 501
	default:
		Log(Database, "Unexpected database error", "error", err)
		return "Internal server error", 500
	}
}

func GqlError(msg, detail string, code int, ctx context.Context) *gqlerror.Error {
	return &gqlerror.Error{
		Path:    graphql.GetPath(ctx),
		Message: msg + ": " + detail,
		Extensions: map[string]any{
			"code": code,
		},
	}
}
