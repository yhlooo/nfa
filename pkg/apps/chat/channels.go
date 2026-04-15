package chat

import (
	"context"
	"fmt"

	"github.com/coder/acp-go-sdk"

	"github.com/yhlooo/nfa/pkg/agents"
	"github.com/yhlooo/nfa/pkg/channels"
)

const (
	channelIDMetaKey = "channelID"
)

// handleChannel 处理信道
func (chat *Chat) handleChannel(ctx context.Context, id int, ch channels.Channel) {
	for msg := range ch.Receive() {
		meta := map[string]any{
			agents.MetaKeyCurrentModels: chat.curModels,
			channelIDMetaKey:            id,
		}
		for k, v := range msg.Meta {
			meta[k] = v
		}

		req := acp.PromptRequest{
			SessionId: chat.sessionID,
			Meta:      meta,
			Prompt:    msg.Prompt,
		}
		chat.p.Send(req)
		resp, err := chat.agent.Prompt(chat.ctx, req)
		if err != nil {
			chat.p.Send(fmt.Errorf("new prompt error: %w", err))
		}
		chat.p.Send(resp)
		if err := ch.Send(ctx, meta, nil, true); err != nil {
			chat.logger.Error(err, "send notification to channel error")
		}
	}
	if err := ch.Err(); err != nil {
		chat.p.Send(err)
	}
}
