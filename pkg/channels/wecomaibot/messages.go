package wecomaibot

// 参考 https://developer.work.weixin.qq.com/document/path/101463

const (
	// CmdSubscribe 开始订阅
	CmdSubscribe = "aibot_subscribe"
	// CmdRespondWelcome 回复欢迎消息
	CmdRespondWelcome = "aibot_respond_welcome_msg"
	// CmdRespondMessage 回复消息
	CmdRespondMessage = "aibot_respond_msg"
	// CmdRespondUpdateCard 回复更新卡片
	CmdRespondUpdateCard = "aibot_respond_update_msg"
	// CmdSendMessage 主动推送消息
	CmdSendMessage = "aibot_send_msg"
	// CmdPing 发送心跳
	CmdPing = "ping"

	// CmdMessageCallback 消息回调
	CmdMessageCallback = "aibot_msg_callback"
	// CmdEventCallback 事件回调
	CmdEventCallback = "aibot_event_callback"
)

// MessageMeta 消息元信息
type MessageMeta struct {
	Cmd          string  `json:"cmd,omitempty"`
	ErrorCode    int     `json:"errcode,omitempty"`
	ErrorMessage string  `json:"errmsg,omitempty"`
	Headers      Headers `json:"headers,omitempty"`
}

// ToRequestMeta 转为请求元信息
func (meta MessageMeta) ToRequestMeta() RequestMeta {
	return RequestMeta{
		Cmd:     meta.Cmd,
		Headers: meta.Headers,
	}
}

// ToResponse 转为响应
func (meta MessageMeta) ToResponse() Response {
	return Response{
		ErrorCode:    meta.ErrorCode,
		ErrorMessage: meta.ErrorMessage,
		Headers:      meta.Headers,
	}
}

// RequestMeta 请求元信息
type RequestMeta struct {
	Cmd     string  `json:"cmd"`
	Headers Headers `json:"headers,omitempty"`
}

// GetMeta 获取请求元信息
func (meta RequestMeta) GetMeta() RequestMeta {
	return meta
}

// Request 请求
type Request interface {
	// GetMeta 获取请求元信息
	GetMeta() RequestMeta
}

// Response 响应
type Response struct {
	ErrorCode    int     `json:"errcode"`
	ErrorMessage string  `json:"errmsg"`
	Headers      Headers `json:"headers,omitempty"`
}

// Headers 请求、响应头
type Headers struct {
	// 请求唯一标识，用于关联请求和响应
	RequestID string `json:"req_id"`
}

// SubscribeRequest 订阅请求
type SubscribeRequest struct {
	RequestMeta
	Body SubscribeRequestBody `json:"body"`
}

// SubscribeRequestBody 订阅请求体
type SubscribeRequestBody struct {
	BotID  string `json:"bot_id"`
	Secret string `json:"secret"`
}

// MessageCallbackRequest 消息回调请求
type MessageCallbackRequest struct {
	RequestMeta
	Body MessageCallbackRequestBody `json:"body"`
}

// MessageCallbackRequestBody 消息回调请求体
type MessageCallbackRequestBody struct {
	// 本次回调的唯一性标志，用于事件排重
	MsgID string `json:"msgid"`
	// 智能机器人 BotID
	AIBotID string `json:"aibotid"`
	// 会话 ID，仅群聊类型时返回
	ChatID string `json:"chatid"`
	// 会话类型，single 单聊 group 群聊
	ChatType ChatType `json:"chattype"`
	// 发送人
	From User `json:"from"`

	// 消息内容
	MessageContent
	// 引用的消息内容
	Quote *MessageContent `json:"quote,omitempty"`
}

// ChatType 对话类型
type ChatType string

const (
	// ChatTypeGroup 群聊
	ChatTypeGroup ChatType = "group"
	// ChatTypeSingle 单聊
	ChatTypeSingle ChatType = "single"
)

// MessageType 消息内容类型
type MessageType string

const (
	TextMessage  MessageType = "text"
	ImageMessage MessageType = "image"
	FileMessage  MessageType = "file"
	VoiceMessage MessageType = "voice"
	VideoMessage MessageType = "video"
	MixedMessage MessageType = "mixed"

	EventMessage MessageType = "event"

	StreamMessage MessageType = "stream"
)

// MessageContent 消息内容
type MessageContent struct {
	// 消息类型
	MsgType MessageType `json:"msgtype"`

	// 文本消息内容
	Text *TextMessageContent `json:"text,omitempty"`
	// 图片消息内容
	Image *FileMessageContent `json:"image,omitempty"`
	// 文件消息内容
	File *FileMessageContent `json:"file,omitempty"`
	// 音频消息内容（语音转文字内容）
	Voice *TextMessageContent `json:"voice,omitempty"`
	// 视频消息内容
	Video *FileMessageContent `json:"video,omitempty"`
	// 混合消息内容
	Mixed *MixedMessageContent `json:"mixed,omitempty"`
}

// TextMessageContent 文本消息内容
type TextMessageContent struct {
	Content string `json:"content"`
}

// FileMessageContent 文件消息内容
type FileMessageContent struct {
	URL string `json:"url"`
}

// MixedMessageContent 混合消息内容
type MixedMessageContent struct {
	MsgItem []MessageContent `json:"msg_item"`
}

// EventCallbackRequest 事件回调请求
type EventCallbackRequest struct {
	RequestMeta
	Body EventCallbackRequestBody `json:"body"`
}

// EventCallbackRequestBody 事件回调请求体
type EventCallbackRequestBody struct {
	// 本次回调的唯一性标志，用于事件排重
	MsgID string `json:"msgid"`
	// 智能机器人 BotID
	AIBotID string `json:"aibotid"`
	// 发送人
	From User `json:"from"`
	// 事件发生时间（ UNIX 时间戳）
	CreateTime int64 `json:"create_time"`
	// 消息类型（固定为 event ）
	MsgType MessageType `json:"msgtype"`
	// 事件内容
	Event EventMessageContent `json:"event"`
}

// EventMessageContent 事件消息内容
type EventMessageContent struct {
	EventType string `json:"eventtype"`
}

// EventType 事件类型
type EventType string

const (
	// EnterChatEvent 用户当天首次进入机器人单聊会话
	EnterChatEvent EventType = "enter_chat"
	// TemplateCardEvent 用户点击模板卡片按钮
	TemplateCardEvent EventType = "template_card_event"
	// FeedbackEvent 用户反馈事件
	FeedbackEvent EventType = "feedback_event"
	// DisconnectedEvent 服务端主动断开连接事件
	DisconnectedEvent EventType = "disconnected_event"
)

type User struct {
	UserID string `json:"userid"`
}

// RespondMessageRequest 回复消息请求
type RespondMessageRequest struct {
	RequestMeta
	Body RespondMessageRequestBody `json:"body"`
}

// RespondMessageRequestBody 回复消息请求体
type RespondMessageRequestBody struct {
	// 消息类型（固定为 stream ）
	MsgType MessageType `json:"msgtype"`
	// 流式消息内容
	Stream *StreamMessageContent `json:"stream"`
}

// StreamMessageContent 流式消息内容
type StreamMessageContent struct {
	// 流 ID
	ID string `json:"id"`
	// 流是否结束
	Finish bool `json:"finish"`
	// 内容
	Content string `json:"content"`
	// 反馈
	Feedback Feedback `json:"feedback,omitempty"`
}

// Feedback 反馈
type Feedback struct {
	ID string `json:"id"`
}
