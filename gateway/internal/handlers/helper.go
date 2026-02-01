package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type apiError struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

func writeJSON(w http.ResponseWriter, statusCode int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(v)
}

func httpStatusFromGRPC(code codes.Code) int {
	switch code {
	case codes.InvalidArgument, codes.OutOfRange, codes.FailedPrecondition:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists, codes.Aborted:
		return http.StatusConflict
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Canceled:
		return 499
	default:
		return http.StatusInternalServerError
	}
}

func writeGRPCError(w http.ResponseWriter, err error) {
	if errors.Is(err, context.DeadlineExceeded) {
		writeJSON(w, http.StatusGatewayTimeout, apiError{Error: "deadline exceeded", Code: "DEADLINE_EXCEEDED"})
		return
	}
	if errors.Is(err, context.Canceled) {
		writeJSON(w, 499, apiError{Error: "request canceled", Code: "CANCELED"})
		return
	}

	st, ok := status.FromError(err)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, apiError{Error: "internal error"})
		return
	}

	writeJSON(w, httpStatusFromGRPC(st.Code()), apiError{
		Error: st.Message(),
		Code:  st.Code().String(),
	})
}
