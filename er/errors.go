package er

import (
	"context"
	"fmt"
	kitContext "git.jetbrains.space/orbi/fcsd/kit/context"
	"github.com/pkg/errors"
	"reflect"
)

// FF specifies list of fields
type FF map[string]interface{}

// AppError specifies application error object
type AppError struct {
	error
	grpcStatus *uint32
	httpStatus *uint32
	code       string
	fields     FF
}

// AppErrBuilder allows building AppError object
type AppErrBuilder interface {
	// C attaches a request context to AppError
	C(ctx context.Context) AppErrBuilder
	// F attaches additional fields to AppError object
	// if type of passed field isn't valid, it's just silently ignored
	F(fields FF) AppErrBuilder
	// GrpcSt attaches gRPC status
	// when converting to grpc error it will be checked and if populated, corresponding grpc status is set
	GrpcSt(status uint32) AppErrBuilder
	// HttpSt attaches HTTP status
	// it gives some hint to API gateway layer what HTTP status to return client
	HttpSt(status uint32) AppErrBuilder
	// Err builds error with all specified attributes
	Err() error
}

// appErrBuildImpl implements AppErrBuilder
type appErrBuildImpl struct {
	// app error
	appErr *AppError
}

func (b *appErrBuildImpl) C(ctx context.Context) AppErrBuilder {
	if rCtx, ok := kitContext.Request(ctx); ok {
		b.F(FF{"ctx": rCtx.ToMap()})
	}
	return b
}

func (b *appErrBuildImpl) F(fields FF) AppErrBuilder {
	ff := make(FF, len(b.appErr.fields)+len(fields))
	for k, v := range b.appErr.fields {
		ff[k] = v
	}
	for k, v := range fields {
		if t := reflect.TypeOf(v); t != nil {
			switch {
			case t.Kind() == reflect.Func, t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Func:
				continue
			}
		}
		ff[k] = v

	}
	b.appErr.fields = ff
	return b
}

func (b *appErrBuildImpl) GrpcSt(status uint32) AppErrBuilder {
	b.appErr.grpcStatus = &status
	return b
}

func (b *appErrBuildImpl) HttpSt(status uint32) AppErrBuilder {
	b.appErr.httpStatus = &status
	return b
}

func (b *appErrBuildImpl) Err() error {
	return b.appErr
}

// WithBuilder creates a new AppErrBuilder and default AppError object
func WithBuilder(code string, format string, args ...interface{}) AppErrBuilder {
	b := &appErrBuildImpl{
		appErr: newAppErr(code, format, args...),
	}
	return b
}

// WrapWithBuilder wraps error and returns builder
func WrapWithBuilder(cause error, code string, format string, args ...interface{}) AppErrBuilder {
	b := &appErrBuildImpl{
		appErr: wrap(cause, code, format, args...),
	}
	return b
}

// newAppErr creates a new AppError
func newAppErr(code string, format string, args ...interface{}) *AppError {
	return &AppError{
		error:  errors.Errorf(format, args...),
		code:   code,
		fields: make(FF),
	}
}

// New creates a new AppError and returns error interface
func New(code string, format string, args ...interface{}) error {
	return newAppErr(code, format, args...)
}

// Error returns default error message
func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.code, e.error)
}

// WithStack return error message with stack trace attached
// if you need split fields, assert to *AppError
func (e *AppError) WithStack() string {
	return fmt.Sprintf("%s: %+v", e.code, e.error)
}

func (e *AppError) WithStackErr() error {
	return &withStackAppErr{AppError: e}
}

// Code returns error code
func (e *AppError) Code() string {
	return e.code
}

// Message returns error message
func (e *AppError) Message() string {
	if e.error != nil {
		return e.error.Error()
	} else {
		return ""
	}
}

// Fields returns fields
func (e *AppError) Fields() FF {
	return e.fields
}

// Fields returns fields
func (e *AppError) GrpcStatus() *uint32 {
	return e.grpcStatus
}

// Wrap wraps error to a AppError object
func Wrap(cause error, code string, format string, args ...interface{}) error {
	return wrap(cause, code, format, args...)
}

func wrap(cause error, code string, format string, args ...interface{}) *AppError {
	return &AppError{
		error:  errors.Wrapf(cause, format, args...),
		code:   code,
		fields: make(FF),
	}
}

// Is checks if error interface is asserted to *AppError
// if true, it returns *AppError
func Is(e error) (*AppError, bool) {
	appErr, ok := e.(*AppError)
	return appErr, ok
}

type withStackAppErr struct {
	*AppError
}

func (s *withStackAppErr) Error() string {
	return s.AppError.WithStack()
}
