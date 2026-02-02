package mapper

import (
	"errors"

	"github.com/ChernykhITMO/order-processing-platform/orders/internal/domain"
	"google.golang.org/grpc/codes"
)

func MapDomainError(err error) (codes.Code, string) {
	if err == nil {
		return codes.OK, ""
	}

	switch {
	case errors.Is(err, domain.ErrInvalidOrderID),
		errors.Is(err, domain.ErrInvalidUserID),
		errors.Is(err, domain.ErrInvalidProductID),
		errors.Is(err, domain.ErrInvalidQuantity),
		errors.Is(err, domain.ErrInvalidPrice),
		errors.Is(err, domain.ErrInvalidItems):
		return codes.InvalidArgument, err.Error()
	case errors.Is(err, domain.ErrOrderNotFound):
		return codes.NotFound, err.Error()
	default:
		return codes.Internal, "internal error"
	}
}
