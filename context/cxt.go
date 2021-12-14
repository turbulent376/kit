package context

import (
	"context"
	"encoding/json"

	"git.jetbrains.space/orbi/fcsd/kit/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
)

const (
	CLIENT_TYPE_REST   = "rest"
	CLIENT_TYPE_TEST   = "test"
	CLIENT_TYPE_JOB    = "job"
	CLIENT_TYPE_QUEUE  = "queue"
	CLIENT_TYPE_WS     = "ws"
	CLIENT_TYPE_WEBRTC = "webrtc"
)

type RequestContext struct {
	// request ID
	Rid string `json:"_ctx.rid"`
	// session ID
	Sid string `json:"_ctx.sid"`
	// user ID
	Uid string `json:"_ctx.uid"`
	// user's session ID
	Usid string `json:"_ctx.usid"`
	// username
	Un string `json:"_ctx.un"`
	// chat user id
	Cid string `json:"_ctx.cid"`
	// client type
	Cl string `json:"_ctx.cl"`
}

type requestContextKey struct{}

func NewRequestCtx() *RequestContext {
	return &RequestContext{}
}

func (r *RequestContext) GetRequestId() string {
	return r.Rid
}

func (r *RequestContext) GetSessionId() string {
	return r.Sid
}

func (r *RequestContext) GetUserId() string {
	return r.Uid
}

func (r *RequestContext) GetUserSessionId() string {
	return r.Usid
}

func (r *RequestContext) GetChatUserId() string {
	return r.Cid
}

func (r *RequestContext) GetClientType() string {
	return r.Cl
}

func (r *RequestContext) GetUsername() string {
	return r.Un
}

func (r *RequestContext) Empty() *RequestContext {

	return &RequestContext{
		Rid:  "",
		Sid:  "",
		Uid:  "",
		Usid: "",
		Un:   "",
		Cid:  "",
		Cl:   "none",
	}
}

func (r *RequestContext) WithRequestId(requestId string) *RequestContext {
	r.Rid = requestId
	return r
}

func (r *RequestContext) WithNewRequestId() *RequestContext {
	r.Rid = utils.NewId()
	return r
}

func (r *RequestContext) WithSessionId(sessionId string) *RequestContext {
	r.Sid = sessionId
	return r
}

func (r *RequestContext) WithChatUserId(chatUserId string) *RequestContext {
	r.Cid = chatUserId
	return r
}

func (r *RequestContext) Rest() *RequestContext {
	r.Cl = CLIENT_TYPE_REST
	return r
}

func (r *RequestContext) Webrtc() *RequestContext {
	r.Cl = CLIENT_TYPE_WEBRTC
	return r
}

func (r *RequestContext) Test() *RequestContext {
	r.Cl = CLIENT_TYPE_TEST
	return r
}

func (r *RequestContext) Job() *RequestContext {
	r.Cl = CLIENT_TYPE_JOB
	return r
}

func (r *RequestContext) Queue() *RequestContext {
	r.Cl = CLIENT_TYPE_QUEUE
	return r
}

func (r *RequestContext) Ws() *RequestContext {
	r.Cl = CLIENT_TYPE_WS
	return r
}

func (r *RequestContext) Client(client string) *RequestContext {
	r.Cl = client
	return r
}

func (r *RequestContext) WithUser(userId, username string) *RequestContext {
	r.Uid = userId
	r.Un = username
	return r
}

func (r *RequestContext) WithUserSession(sessionId string) *RequestContext {
	r.Usid = sessionId
	return r
}

func (r *RequestContext) ToContext(parent context.Context) context.Context {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithValue(parent, requestContextKey{}, r)
}

func (r *RequestContext) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"_ctx.rid":  r.Rid,
		"_ctx.sid":  r.Sid,
		"_ctx.uid":  r.Uid,
		"_ctx.usid": r.Usid,
		"_ctx.un":   r.Un,
		"_ctx.cid":  r.Cid,
		"_ctx.cl":   r.Cl,
	}
}

func Request(context context.Context) (*RequestContext, bool) {
	if r, ok := context.Value(requestContextKey{}).(*RequestContext); ok {
		return r, true
	}
	return &RequestContext{}, false
}

func MustRequest(context context.Context) (*RequestContext, error) {
	if r, ok := context.Value(requestContextKey{}).(*RequestContext); ok {
		return r, nil
	}
	return &RequestContext{}, errors.New("context is invalid")
}

func FromMap(ctx context.Context, mp map[string]interface{}) (context.Context, error) {
	var r *RequestContext
	err := mapstructure.Decode(mp, &r)
	if err != nil {
		return nil, err
	}
	return r.ToContext(ctx), nil
}

func FromGrpcMD(ctx context.Context, md metadata.MD) context.Context {

	if rqb, ok := md["rq-bin"]; ok {
		if len(rqb) > 0 {
			rm := []byte(rqb[0])
			rq := &RequestContext{}
			_ = json.Unmarshal(rm, rq)
			return context.WithValue(ctx, requestContextKey{}, rq)
		}
	}
	return ctx
}

func FromContextToGrpcMD(ctx context.Context) (metadata.MD, bool) {
	if r, ok := Request(ctx); ok {
		rm, _ := json.Marshal(*r)
		return metadata.Pairs("rq-bin", string(rm)), true
	}
	return metadata.Pairs(), false
}
