package http

import (
	"context"
	"net/http"

	"git.jetbrains.space/orbi/fcsd/kit/er"
)

const (
	ErrCodeHttpSrvListen                     = "HTTP-001"
	ErrCodeDecodeRequest                     = "HTTP-002"
	ErrCodeHttpUrlVar                        = "HTTP-003"
	ErrCodeHttpCurrentUser                   = "HTTP-004"
	ErrCodeHttpUrlVarEmpty                   = "HTTP-005"
	ErrCodeHttpUrlFormVarEmpty               = "HTTP-006"
	ErrCodeHttpUrlFormVarNotInt              = "HTTP-007"
	ErrCodeHttpUrlFormVarNotTime             = "HTTP-008"
	ErrCodeHttpMultipartParseForm            = "HTTP-009"
	ErrCodeHttpMultipartEmptyContent         = "HTTP-010"
	ErrCodeHttpMultipartNotMultipart         = "HTTP-011"
	ErrCodeHttpMultipartParseMediaType       = "HTTP-012"
	ErrCodeHttpMultipartWrongMediaType       = "HTTP-013"
	ErrCodeHttpMultipartMissingBoundary      = "HTTP-014"
	ErrCodeHttpMultipartEofReached           = "HTTP-015"
	ErrCodeHttpMultipartNext                 = "HTTP-016"
	ErrCodeHttpMultipartFormNameFileExpected = "HTTP-017"
	ErrCodeHttpMultipartFilename             = "HTTP-018"
)

var (
	ErrHttpSrvListen     = func(cause error) error { return er.WrapWithBuilder(cause, ErrCodeHttpSrvListen, "").Err() }
	ErrHttpDecodeRequest = func(cause error, ctx context.Context) error {
		return er.WrapWithBuilder(cause, ErrCodeDecodeRequest, "invalid request").C(ctx).HttpSt(http.StatusBadRequest).Err()
	}
	ErrHttpUrlVar = func(ctx context.Context, v string) error {
		return er.WithBuilder(ErrCodeHttpUrlVar, "invalid or empty URL parameter").F(er.FF{"var": v}).C(ctx).HttpSt(http.StatusBadRequest).Err()
	}
	ErrHttpCurrentUser = func(ctx context.Context) error {
		return er.WithBuilder(ErrCodeHttpCurrentUser, `cannot obtain current user`).C(ctx).HttpSt(http.StatusBadRequest).Err()
	}
	ErrHttpCurrentSession = func(ctx context.Context) error {
		return er.WithBuilder(ErrCodeHttpCurrentUser, `cannot obtain current session`).C(ctx).HttpSt(http.StatusBadRequest).Err()
	}
	ErrHttpUrlVarEmpty = func(ctx context.Context, v string) error {
		return er.WithBuilder(ErrCodeHttpUrlVarEmpty, `URL parameter is empty`).F(er.FF{"var": v}).C(ctx).HttpSt(http.StatusBadRequest).Err()
	}
	ErrHttpUrlFormVarEmpty = func(ctx context.Context, v string) error {
		return er.WithBuilder(ErrCodeHttpUrlFormVarEmpty, `URL form value is empty`).F(er.FF{"var": v}).C(ctx).HttpSt(http.StatusBadRequest).Err()
	}
	ErrHttpUrlFormVarNotInt = func(cause error, ctx context.Context, v string) error {
		return er.WrapWithBuilder(cause, ErrCodeHttpUrlFormVarNotInt, "form value must be of int type").C(ctx).HttpSt(http.StatusBadRequest).Err()
	}
	ErrHttpUrlFormVarNotTime = func(cause error, ctx context.Context, v string) error {
		return er.WrapWithBuilder(cause, ErrCodeHttpUrlFormVarNotTime, "form value must be of time type in RFC-3339 format").C(ctx).HttpSt(http.StatusBadRequest).Err()
	}
	ErrHttpMultipartParseForm = func(cause error, ctx context.Context) error {
		return er.WrapWithBuilder(cause, ErrCodeHttpMultipartParseForm, "parse multipart form").C(ctx).HttpSt(http.StatusBadRequest).Err()
	}
	ErrHttpMultipartEmptyContent = func(ctx context.Context) error {
		return er.WithBuilder(ErrCodeHttpMultipartEmptyContent, `content is empty`).C(ctx).HttpSt(http.StatusBadRequest).Err()
	}
	ErrHttpMultipartNotMultipart = func(ctx context.Context) error {
		return er.WithBuilder(ErrCodeHttpMultipartNotMultipart, `content isn't multipart`).C(ctx).HttpSt(http.StatusBadRequest).Err()
	}
	ErrHttpMultipartParseMediaType = func(cause error, ctx context.Context) error {
		return er.WrapWithBuilder(cause, ErrCodeHttpMultipartParseMediaType, "parse media type").C(ctx).HttpSt(http.StatusBadRequest).Err()
	}
	ErrHttpMultipartWrongMediaType = func(ctx context.Context, mt string) error {
		return er.WithBuilder(ErrCodeHttpMultipartWrongMediaType, `wrong media type %s`, mt).C(ctx).HttpSt(http.StatusBadRequest).Err()
	}
	ErrHttpMultipartMissingBoundary = func(ctx context.Context) error {
		return er.WithBuilder(ErrCodeHttpMultipartMissingBoundary, `missing boundary`).C(ctx).HttpSt(http.StatusBadRequest).Err()
	}
	ErrHttpMultipartEofReached = func(ctx context.Context) error {
		return er.WithBuilder(ErrCodeHttpMultipartEofReached, `no parts found`).C(ctx).HttpSt(http.StatusBadRequest).Err()
	}
	ErrHttpMultipartNext = func(cause error, ctx context.Context) error {
		return er.WrapWithBuilder(cause, ErrCodeHttpMultipartNext, "reading part").C(ctx).HttpSt(http.StatusBadRequest).Err()
	}
	ErrHttpMultipartFormNameFileExpected = func(ctx context.Context) error {
		return er.WithBuilder(ErrCodeHttpMultipartFormNameFileExpected, `correct part must have name="file" param`).C(ctx).HttpSt(http.StatusBadRequest).Err()
	}
	ErrHttpMultipartFilename = func(ctx context.Context) error {
		return er.WithBuilder(ErrCodeHttpMultipartFilename, `filename is empty`).C(ctx).HttpSt(http.StatusBadRequest).Err()
	}
)
