package ctype

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type MsgType int8

const (
	TextMsgType MsgType = iota + 1
	ImageMsgType
	VideoMsgType
	FileMsgType
	VoiceMsgType
	VoiceCallMsgType
	VideoCallMsgType
	WithdrawMsgType
	ReplyMsgType
	QuoteMsgType
	AtMsgType
	TipMsgType
	FriendOnlineMsgType
	ImageTextMsgType
)

type Msg struct {
	Type            MsgType          `json:"type"`                      // 消息类型 和msgType一模一样
	TextMsg         *TextMsg         `json:"textMsg,omitempty"`         // 文本消息
	ImageMsg        *ImageMsg        `json:"imageMsg,omitempty"`        // 图片消息
	VideoMsg        *VideoMsg        `json:"videoMsg,omitempty"`        // 视频消息
	FileMsg         *FileMsg         `json:"fileMsg,omitempty"`         // 文件消息
	VoiceMsg        *VoiceMsg        `json:"voiceMsg,omitempty"`        // 语音消息
	VoiceCallMsg    *VoiceCallMsg    `json:"voiceCallMsg,omitempty"`    // 语音通话
	VideoCallMsg    *VideoCallMsg    `json:"videoCallMsg,omitempty"`    // 视频通话
	WithdrawMsg     *WithdrawMsg     `json:"withdrawMsg,omitempty"`     // 撤回消息
	ReplyMsg        *ReplyMsg        `json:"replyMsg,omitempty"`        // 回复消息
	QuoteMsg        *QuoteMsg        `json:"quoteMsg,omitempty"`        // 引用消息
	AtMsg           *AtMsg           `json:"atMsg,omitempty"`           // @用户的消息 群聊才有
	TipMsg          *TipMsg          `json:"tipMsg,omitempty"`          // 提示消息 一般是不入库的
	FriendOnlineMsg *FriendOnlineMsg `json:"friendOnlineMsg,omitempty"` // 好友上线提醒 不入库的
	ImageTextMsg    *ImageTextMsg    `json:"imageTextMsg,omitempty"`    // 图文消息
}

func (msg Msg) MsgPreview() string {
	switch msg.Type {
	case 1:
		var runes = []rune(msg.TextMsg.Content)
		if len(runes) > 30 {
			return string(runes[:30])
		}
		return msg.TextMsg.Content
	case 2:
		return "[图片消息] - " + msg.ImageMsg.Title
	case 3:
		return "[视频消息] - " + msg.VideoMsg.Title
	case 4:
		return "[文件消息] - " + msg.FileMsg.Title
	case 5:
		return "[语音消息]"
	case 6:
		return "[语言通话]"
	case 7:
		return "[视频通话]"
	case 8:
		return "[撤回消息] - " + msg.WithdrawMsg.Content
	case 9:
		return "[回复消息] - " + msg.ReplyMsg.Content
	case 10:
		return "[引用消息] - " + msg.QuoteMsg.Content
	case 11:
		return "[@消息] - " + msg.AtMsg.Content
	case 14:
		return "[图文消息]"
	}
	return "[未知消息]"
}
func (msg Msg) Validate() error {
	switch msg.Type {
	case TextMsgType:
		if msg.TextMsg == nil {
			return errors.New("文本消息不能为空")
		}
		return msg.TextMsg.Validate()
	case ImageMsgType:
		if msg.ImageMsg == nil {
			return errors.New("图片消息不能为空")
		}
		return msg.ImageMsg.Validate()
	case VideoMsgType:
		if msg.VideoMsg == nil {
			return errors.New("视频消息不能为空")
		}
		return msg.VideoMsg.Validate()
	case FileMsgType:
		if msg.FileMsg == nil {
			return errors.New("文件消息不能为空")
		}
		return msg.FileMsg.Validate()
	case ImageTextMsgType:
		if msg.ImageTextMsg == nil {
			return errors.New("图文消息不能为空")
		}
		return msg.ImageTextMsg.Validate()

	}
	return nil
}
func (c *Msg) Scan(val interface{}) error {
	return json.Unmarshal(val.([]byte), c)
}

// Value 入库的数据
func (c Msg) Value() (driver.Value, error) {
	b, err := json.Marshal(c)
	return string(b), err
}

type TextMsg struct {
	Content string `json:"content"`
}

func (t TextMsg) Validate() error {
	if t.Content == "" {
		return errors.New("请输入文本消息的content")
	}
	return nil
}

type ImageMsg struct {
	Title string `json:"title"`
	Src   string `json:"src"`
}

func (t ImageMsg) Validate() error {
	if t.Title == "" {
		return errors.New("请输入图片消息的title")
	}
	if t.Src == "" {
		return errors.New("请输入图片消息的src")
	}
	return nil
}

type VideoMsg struct {
	Title string `json:"title"`
	Src   string `json:"src"`
	Time  int    `json:"time"` // 时长 单位秒
}

func (t VideoMsg) Validate() error {
	if t.Title == "" {
		return errors.New("请输入视频消息的title")
	}
	if t.Src == "" {
		return errors.New("请输入视频消息的src")
	}
	return nil
}

type FileMsg struct {
	Title string `json:"title"`
	Src   string `json:"src"`
	Size  int64  `json:"size"` // 文件大小
	Type  string `json:"type"` // 文件类型 word
}

func (t FileMsg) Validate() error {
	if t.Src == "" {
		return errors.New("请输入文件消息的src")
	}
	return nil
}

type VoiceMsg struct {
	Src  string `json:"src"`
	Time int    `json:"time"` // 时长 单位秒
}

func (t VoiceMsg) Validate() error {

	if t.Src == "" {
		return errors.New("请输入语音消息的src")
	}
	return nil
}

type VoiceCallMsg struct {
	StartTime time.Time `json:"startTime"` // 开始时间
	EndTime   time.Time `json:"endTime"`   // 结束时间
	EndReason int8      `json:"endReason"` // 结束原因 0 发起方挂断 1 接收方挂断  2  网络原因挂断  3 未打通
}
type VideoCallMsg struct {
	StartTime time.Time `json:"startTime"` // 开始时间
	EndTime   time.Time `json:"endTime"`   // 结束时间
	EndReason int8      `json:"endReason"` // 结束原因 0 发起方挂断 1 接收方挂断  2  网络原因挂断  3 未打通
	Flag      int8      `json:"flag"`      // 标识，标识客户端弹框的模式
	Msg       string    `json:"msg"`
	Type      string    `json:"type"`
	Data      any       `json:"data"`
}

// WithdrawMsg 撤回消息
type WithdrawMsg struct {
	Content   string `json:"content"`             // 撤回的提示词
	MsgID     uint   `json:"msgID"`               // 需要撤回的消息id 入参必填
	OriginMsg *Msg   `json:"originMsg,omitempty"` // 原消息  怎么做到，转出去的时候不显示
}
type ReplyMsg struct {
	MsgID           uint      `json:"msgID"`   // 消息id
	Content         string    `json:"content"` // 回复的文本消息，目前只能限制回复文本
	Msg             *Msg      `json:"msg,omitempty"`
	ReplyMsgPreview string    `json:"replyMsgPreview"` // 回复的消息预览
	UserID          uint      `json:"userID"`          // 被回复人的用户id
	UserNickName    string    `json:"userNickName"`    // 被回复人的昵称
	OriginMsgDate   time.Time `json:"originMsgDate"`   // 原消息的时间
}
type QuoteMsg struct {
	MsgID           uint      `json:"msgID"`   // 消息id
	Content         string    `json:"content"` // 回复的文本消息，目前只能限制回复文本
	Msg             *Msg      `json:"msg"`
	UserID          uint      `json:"userID"`          // 被回复人的用户id
	UserNickName    string    `json:"userNickName"`    // 被回复人的昵称
	OriginMsgDate   time.Time `json:"originMsgDate"`   // 原消息的时间
	QuoteMsgPreview string    `json:"quoteMsgPreview"` // 回复的消息预览
}

// AtMsg @消息
type AtMsg struct {
	UserID  uint   `json:"userID"`
	Content string `json:"content"` // 回复的文本消息
	Msg     *Msg   `json:"msg"`
}
type TipMsg struct {
	Status  string `json:"status"`  // 状态  success 成功  error 失败
	Content string `json:"content"` // 提示内容
}
type FriendOnlineMsg struct {
	NickName string `json:"nickName"` // 昵称
	Avatar   string `json:"avatar"`   // 头像
	Content  string `json:"content"`  // 内容
	FriendID uint   `json:"friendID"` // 好友id
}
type ImageTextMsg struct {
	Content string `json:"content"` // 内容
}

func (t ImageTextMsg) Validate() error {
	if t.Content == "" {
		return errors.New("请输入图文消息的content")
	}
	//if safe.SanitizeHTML(t.Content) != t.Content {
	//	return errors.New("图文消息的content不能包含html标签")
	//}
	return nil
}
