package util

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Like status.Code(), but supports wrapped errors.
func StatusCode(err error) codes.Code {
	if err == nil {
		return codes.OK
	}
	var grpcStatus interface{ GRPCStatus() *status.Status }
	code := codes.Unknown
	if errors.As(err, &grpcStatus) {
		code = grpcStatus.GRPCStatus().Code()
	}
	return code
}
