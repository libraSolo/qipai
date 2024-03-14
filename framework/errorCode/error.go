package errorCode

import (
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Error struct {
	Code int
	Err  error
}

func (e *Error) Error() string {
	return e.Err.Error()
}

func NewError(code int, err error) *Error {
	return &Error{
		Code: code,
		Err:  err,
	}
}

func GrpcError(err *Error) error {
	return status.Error(codes.Code(err.Code), err.Err.Error())
}

func ToError(err error) *Error {
	if s, ok := status.FromError(err); ok {
		return NewError(int(s.Code()), errors.New(s.Message()))
	}
	return nil
}
