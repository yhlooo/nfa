package yuanbaobot

// 参考 stories/yuanbao-bot.md

// 连接层 Module
const ModuleConnAccess = "conn_access"

// 业务 Module
const ModuleBiz = "yuanbao_openclaw_proxy"

// 连接层命令
const (
	CmdAuthBind   = "auth-bind"
	CmdPing       = "ping"
	CmdKickout    = "kickout"
	CmdUpdateMeta = "update-meta"
)

// 业务命令
const (
	CmdSendC2CMessage       = "send_c2c_message"
	CmdSendGroupMessage     = "send_group_message"
	CmdQueryGroupInfo       = "query_group_info"
	CmdGetGroupMemberList   = "get_group_member_list"
	CmdSendPrivateHeartbeat = "send_private_heartbeat"
	CmdSendGroupHeartbeat   = "send_group_heartbeat"
	CmdSyncInformation      = "sync_information"
)

// 回调命令
const (
	CallbackC2CSendMsg     = "C2C.CallbackAfterSendMsg"
	CallbackGroupSendMsg   = "Group.CallbackAfterSendMsg"
	CallbackC2CMsgWithdraw = "C2C.CallbackAfterMsgWithDraw"
	CallbackGroupRecallMsg = "Group.CallbackAfterRecallMsg"
)

// 消息类型
const (
	MsgTypeText   = "TIMTextElem"
	MsgTypeImage  = "TIMImageElem"
	MsgTypeFile   = "TIMFileElem"
	MsgTypeSound  = "TIMSoundElem"
	MsgTypeVideo  = "TIMVideoFileElem"
	MsgTypeCustom = "TIMCustomElem"
)

// 不重连的 WebSocket 关闭码
var NoReconnectCloseCodes = map[int]bool{
	4012: true, 4013: true, 4014: true,
	4018: true, 4019: true, 4021: true,
}

// Token 过期的认证错误码
var TokenExpiredCodes = map[int]bool{
	41103: true, // AUTH_TOKEN_INVALID
	41104: true, // AUTH_TOKEN_EXPIRED
	41108: true, // AUTH_TOKEN_FORCED_EXPIRATION
}

// InboundMessageJSON 入站消息 JSON 结构（服务端推送的 inbound_message）
type InboundMessageJSON struct {
	CallbackCommand string               `json:"callback_command"`
	FromAccount     string               `json:"from_account"`
	ToAccount       string               `json:"to_account"`
	SenderNickname  string               `json:"sender_nickname"`
	GroupID         string               `json:"group_id,omitempty"`
	GroupCode       string               `json:"group_code,omitempty"`
	GroupName       string               `json:"group_name,omitempty"`
	MsgSeq          int64                `json:"msg_seq"`
	MsgRandom       int64                `json:"msg_random"`
	MsgTime         int64                `json:"msg_time"`
	MsgKey          string               `json:"msg_key"`
	MsgID           string               `json:"msg_id"`
	MsgBody         []InboundMsgBodyJSON `json:"msg_body"`
	CloudCustomData string               `json:"cloud_custom_data,omitempty"`
	BotOwnerID      string               `json:"bot_owner_id"`
	ClawMsgType     int32                `json:"claw_msg_type"`
	LogExt          *InboundLogExtJSON   `json:"log_ext,omitempty"`
}

// InboundMsgBodyJSON 入站消息体 JSON
type InboundMsgBodyJSON struct {
	MsgType    string                `json:"msg_type"`
	MsgContent InboundMsgContentJSON `json:"msg_content"`
}

// InboundMsgContentJSON 入站消息内容 JSON
type InboundMsgContentJSON struct {
	Text string `json:"text,omitempty"`
}

// InboundLogExtJSON 入站消息日志扩展 JSON
type InboundLogExtJSON struct {
	TraceID string `json:"trace_id,omitempty"`
}
