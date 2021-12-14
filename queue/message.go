package queue

import (
	"context"
	"encoding/json"
	kitCtx "git.jetbrains.space/orbi/fcsd/kit/context"
)

type Message struct {
	Ctx     *kitCtx.RequestContext `json:"ctx"`
	Payload interface{}            `json:"pl"`
}

func Decode(parentCtx context.Context, msg []byte, payload interface{}) (context.Context, error) {

	var m Message

	err := json.Unmarshal(msg, &m)
	if err != nil {
		return nil, ErrQueueMsgUnmarshal(err)
	}

	_, ok := payload.(map[string]interface{})
	// if target type isn't map[string]interface{} try to decode, otherwise it's already it
	if !ok {
		plM, _ := json.Marshal(m.Payload)
		err = json.Unmarshal(plM, &payload)
		if err != nil {
			return nil, ErrQueueMsgUnmarshalPayload(err)
		}
		m.Payload = payload
	} else {
		payload = m.Payload
	}

	if parentCtx == nil {
		parentCtx = context.Background()
	}

	ctx := m.Ctx.ToContext(parentCtx)

	return ctx, nil

}
