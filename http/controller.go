package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	kitContext "git.jetbrains.space/orbi/fcsd/kit/context"
	"git.jetbrains.space/orbi/fcsd/kit/er"
	"github.com/gorilla/mux"
)

// Error is a HTTP error object returning to clients in case of error
type Error struct {
	Code    string                 `json:"code,omitempty"`    // Code is error code provided by error producer
	Message string                 `json:"message"`           // Message is error description
	Details map[string]interface{} `json:"details,omitempty"` // Details is additional info provided by error producer
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s:%s", e.Code, e.Message)
}

const (
	Me = "me" // Me can be used in URL whenever userId is expected. When encountered, userId from the session context is used
)

var EmptyOkResponse = struct {
	Status string `json:"status"`
}{
	Status: "OK",
}

type BaseController struct{}

var MediaContentTypes = [...]string{
	"image/jpeg",
	"image/png",
	"image/bmp",
	"image/gif",
	"image/tiff",
	"video/avi",
	"video/mpeg",
	"video/mp4",
	"audio/mpeg",
	"audio/wav",
}

type ResponseContentOpts struct {
	Filename     string
	ContentType  string
	ContentSize  int
	Download     bool
	ModifiedTime time.Time
}

func (c *BaseController) RespondContent(w http.ResponseWriter, r *http.Request, opts ResponseContentOpts, file []byte) {

	w.Header().Set("Cache-Control", "private, no-cache")

	if opts.ContentSize > 0 {
		contentSizeStr := strconv.Itoa(opts.ContentSize)
		w.Header().Set("Content-Length", contentSizeStr)
	}

	if opts.ContentType == "" {
		opts.ContentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", opts.ContentType)

	if !opts.Download {
		isMedia := false
		for _, mct := range MediaContentTypes {
			if strings.HasPrefix(opts.ContentType, mct) {
				isMedia = true
				break
			}
		}
		opts.Download = !isMedia
	}

	if opts.Download {
		w.Header().Set("Content-Disposition", "attachment;filename=\""+opts.Filename+"\"; filename*=UTF-8''"+opts.Filename)
	} else {
		w.Header().Set("Content-Disposition", "inline;filename=\""+opts.Filename+"\"; filename*=UTF-8''"+opts.Filename)
	}

	http.ServeContent(w, r, opts.Filename, opts.ModifiedTime, bytes.NewReader(file))

}

// GetUploadFileMultipartContent it parse body for multipart content disposition
// it expects the only one part with the following structure:
//-----------------------------4562559108110960722260982980
//Content-Disposition: form-data; name="files"; filename="my-file.jpg"
//Content-Type: image/jpeg
//....
//.....
func (c *BaseController) GetUploadFileMultipartContent(ctx context.Context, r *http.Request) (io.Reader, string, error) {

	// parse form
	if r.Form == nil {
		err := r.ParseForm()
		if err != nil {
			return nil, "", ErrHttpMultipartParseForm(err, ctx)
		}
	}
	if r.ContentLength == 0 {
		return nil, "", ErrHttpMultipartEmptyContent(ctx)
	}

	// get content type from header
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		return nil, "", ErrHttpMultipartNotMultipart(ctx)
	}

	// parse mime type
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, "", ErrHttpMultipartParseMediaType(err, ctx)
	}
	if mediaType != "multipart/form-data" {
		return nil, "", ErrHttpMultipartWrongMediaType(ctx, mediaType)
	}

	// identify boundary
	boundary, ok := params["boundary"]
	if !ok {
		return nil, "", ErrHttpMultipartMissingBoundary(ctx)
	}

	// create a new reader
	mr := multipart.NewReader(r.Body, boundary)

	// go through all parts
	for {

		// take next part
		part, err := mr.NextPart()
		if err != nil {
			if err == io.EOF {
				// if we get here, we haven't found any useful parts, so it's wrong format
				return nil, "", ErrHttpMultipartEofReached(ctx)
			} else {
				return nil, "", ErrHttpMultipartNext(err, ctx)
			}
		}

		// check found part
		if part.FormName() == "file" {
			filename := part.FileName()
			if filename == "" {
				return nil, "", ErrHttpMultipartFilename(ctx)
			}
			// return first part
			return part, filename, nil
		} else {
			return nil, "", ErrHttpMultipartFormNameFileExpected(ctx)
		}

	}
}

func (c *BaseController) RespondJson(w http.ResponseWriter, httpStatus int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_, _ = w.Write(response)
}

func (c *BaseController) RespondError(w http.ResponseWriter, err error) {

	httpErr := &Error{}
	httpStatus := http.StatusInternalServerError

	// check if this is an app error
	if appErr, ok := er.Is(err); ok {
		httpErr.Code = appErr.Code()
		httpErr.Message = appErr.Message()
		httpErr.Details = appErr.Fields()
	} else {
		httpErr.Message = err.Error()
	}
	c.RespondJson(w, httpStatus, httpErr)
}

func (c *BaseController) RespondWithStatus(w http.ResponseWriter, status int, payload interface{}) {
	c.RespondJson(w, status, payload)
}

func (c *BaseController) RespondOK(w http.ResponseWriter, payload interface{}) {
	c.RespondJson(w, http.StatusOK, payload)
}

func (c *BaseController) DecodeRequest(r *http.Request, ctx context.Context, body interface{}) error {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(body); err != nil {
		return ErrHttpDecodeRequest(err, ctx)
	}
	return nil
}

func (c *BaseController) Var(r *http.Request, ctx context.Context, varName string, allowEmpty bool) (string, error) {
	if val, ok := mux.Vars(r)[varName]; ok {
		if !allowEmpty && val == "" {
			return "", ErrHttpUrlVarEmpty(ctx, varName)
		}
		return val, nil
	} else {
		return "", ErrHttpUrlVar(ctx, varName)
	}
}


func (c *BaseController) CurrentUser(ctx context.Context) (uid string, usid string, err error) {
	if rCtx, ok := kitContext.Request(ctx); ok {
		if rCtx.Uid != "" && rCtx.Usid != "" {
			return rCtx.Uid, rCtx.Usid, nil
		} else {
			return "", "", ErrHttpCurrentSession(ctx)
		}
	} else {
		return "", "", ErrHttpCurrentSession(ctx)
	}
}

func (c *BaseController) UserIdVar(r *http.Request, ctx context.Context, varName string) (string, error) {
	val, err := c.Var(r, ctx, varName, false)
	if err != nil {
		return "", err
	}
	if val == Me {
		if uid, _, err := c.CurrentUser(ctx); err != nil {
			return "", err
		} else {
			return uid, nil
		}
	}
	return val, nil
}

func (c *BaseController) FormVal(r *http.Request, ctx context.Context, name string, allowEmpty bool) (string, error) {
	val := r.FormValue(name)
	if !allowEmpty && val == "" {
		return "", ErrHttpUrlFormVarEmpty(ctx, name)
	}
	return val, nil
}

func (c *BaseController) FormValInt(r *http.Request, ctx context.Context, name string, allowEmpty bool) (*int, error) {
	valStr, err := c.FormVal(r, ctx, name, allowEmpty)
	if err != nil {
		return nil, err
	}
	if allowEmpty && valStr == "" {
		return nil, nil
	}
	valInt, err := strconv.Atoi(valStr)
	if err != nil {
		return nil, ErrHttpUrlFormVarNotInt(err, ctx, name)
	}
	return &valInt, nil
}

// FormValTime parses URL form value and checks for time in RFC3339 format(UTC)
func (c *BaseController) FormValTime(r *http.Request, ctx context.Context, name string, allowEmpty bool) (*time.Time, error) {
	valStr, err := c.FormVal(r, ctx, name, allowEmpty)
	if err != nil {
		return nil, err
	}
	if allowEmpty && valStr == "" {
		return nil, nil
	}
	valTime, err := time.Parse(time.RFC3339, valStr)
	if err != nil {
		return nil, ErrHttpUrlFormVarNotTime(err, ctx, name)
	}
	return &valTime, nil
}
