package handler

import (
	"context"
	"encoding/json"
	"fim/common/models/ctype"
	"fim/common/response"
	"fim/common/service/redis_service"
	"fim/fim_file/file_rpc/files"
	"fim/fim_group/group_api/internal/svc"
	"fim/fim_group/group_api/internal/types"
	"fim/fim_group/group_models"
	"fim/fim_user/user_rpc/types/user_rpc"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type UserWsInfo struct {
	UserINfo    ctype.UserInfo             //用户信息
	WsClientMap map[string]*websocket.Conn //用户管理所有ws连接
}

var UserOnlineWsMap = map[uint]*UserWsInfo{}

type ChatRequest struct {
	GroupID uint      `json:"group_id"` //群组id
	Msg     ctype.Msg `json:"msg"`      //消息内容
}
type ChatResponse struct {
	GroupID        uint          `json:"groupID"`
	UserID         uint          `json:"userID"`
	UserNickname   string        `json:"userNickname"`
	UserAvatar     string        `json:"userAvatar"`
	Msg            ctype.Msg     `json:"msg"`
	ID             uint          `json:"id"`
	MsgType        ctype.MsgType `json:"msgType"`
	CreatedAt      time.Time     `json:"createdAt"`
	IsMe           bool          `json:"isMe"`
	MemberNickname string        `json:"memberNickname"` // 群好友备注
	MsgPreview     string        `json:"msgPreview"`
}

func groupChatHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GroupChatRequest
		if err := httpx.Parse(r, &req); err != nil {
			response.Response(r, w, nil, err)
			return
		}

		var upGrader = websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
		conn, err := upGrader.Upgrade(w, r, nil)
		if err != nil {
			logx.Error(err)
			response.Response(r, w, nil, err)
			return
		}
		addr := conn.RemoteAddr().String()
		logx.Infof("用户建立ws连接:%s", addr)
		defer func() {
			conn.Close()
			userWsInfo, ok := UserOnlineWsMap[req.UserID]
			if ok {
				delete(userWsInfo.WsClientMap, addr)
			}
			if userWsInfo != nil && len(userWsInfo.WsClientMap) == 0 {
				delete(UserOnlineWsMap, req.UserID)
			}
		}()
		baseInfoResponse, err := svcCtx.UserRpc.UserBaseInfo(context.Background(), &user_rpc.UserBaseInfoRequest{
			UserId: uint32(req.UserID),
		})
		if err != nil {
			logx.Error(err)
			response.Response(r, w, nil, err)
			return
		}
		userInfo := ctype.UserInfo{
			ID:       req.UserID,
			NickName: baseInfoResponse.NickName,
			Avatar:   baseInfoResponse.Avatar,
		}
		userWsInfo, ok := UserOnlineWsMap[req.UserID]
		if !ok {
			userWsInfo = &UserWsInfo{
				UserINfo: userInfo,
				WsClientMap: map[string]*websocket.Conn{
					addr: conn,
				},
			}
			UserOnlineWsMap[req.UserID] = userWsInfo
		}
		_, ok1 := userWsInfo.WsClientMap[addr]
		if !ok1 {
			UserOnlineWsMap[req.UserID].WsClientMap[addr] = conn
		}
		for {
			_, p, err1 := conn.ReadMessage()
			if err1 != nil {
				fmt.Println(err1)
				break
			}
			var request ChatRequest
			err = json.Unmarshal(p, &request)
			if err != nil {
				logx.Error(err)
				SendTipErrMsg(conn, "参数解析失败")
				continue
			}
			msgValidateErr := request.Msg.Validate()
			if msgValidateErr != nil {
				SendTipErrMsg(conn, msgValidateErr.Error())
				continue
			}
			var member group_models.GroupMemberModel
			err = svcCtx.DB.Preload("GroupModel").Take(&member, "group_id = ? and user_id=?", request.GroupID, req.UserID).Error
			if err != nil {
				SendTipErrMsg(conn, "群组不存在")
				continue
			}
			if member.GroupModel.IsProhibition && member.Role == 3 {
				SendTipErrMsg(conn, "群被禁言")
				continue
			}
			if member.GetProhibitionTime(svcCtx.Redis, svcCtx.DB) != nil {
				SendTipErrMsg(conn, "您已被禁言")
				continue
			}
			switch request.Msg.Type {
			case ctype.FileMsgType:
				nameList := strings.Split(request.Msg.FileMsg.Src, "/")
				if len(nameList) == 0 {
					SendTipErrMsg(conn, "请上传文件")
					continue
				}
				fileID := nameList[len(nameList)-1]
				fileResponse, err3 := svcCtx.FileRpc.FileInfo(context.Background(), &files.FileInfoRequest{
					FileId: fileID,
				})
				if err3 != nil {
					logx.Error(err3)
					SendTipErrMsg(conn, err3.Error())
					continue
				}
				request.Msg.FileMsg.Title = fileResponse.FileName
				request.Msg.FileMsg.Size = fileResponse.FileSize
				request.Msg.FileMsg.Type = fileResponse.FileType
			case ctype.WithdrawMsgType:
				withdrawMsg := request.Msg.WithdrawMsg
				if withdrawMsg == nil {
					SendTipErrMsg(conn, "撤回消息的格式错误")
					continue
				}
				if withdrawMsg.MsgID == 0 {
					SendTipErrMsg(conn, "撤回消息id错误")
					continue
				}
				var groupMsg group_models.GroupMsgModel
				err = svcCtx.DB.Take(&groupMsg, "group_id=? and id =?", request.GroupID, withdrawMsg.MsgID).Error
				if err != nil {
					SendTipErrMsg(conn, "撤回消息不存在")
					continue
				}
				if groupMsg.MsgType == ctype.WithdrawMsgType {
					SendTipErrMsg(conn, "消息已被撤回")
					continue
				}
				if member.Role == 3 {
					if req.UserID != groupMsg.SendUserID {
						SendTipErrMsg(conn, "您没有权限撤回此消息")
						continue
					}
					now := time.Now()
					if now.Sub(groupMsg.CreatedAt) > 2*time.Minute {
						SendTipErrMsg(conn, "消息已超过2分钟，无法撤回")
						continue
					}
				}
				var msgUserRole int8 = 3
				err = svcCtx.DB.Model(group_models.GroupMemberModel{}).
					Where("group_id=? and user_id=?", request.GroupID, groupMsg.SendUserID).
					Select("role").Scan(&msgUserRole).Error
				if member.Role == 2 {
					if msgUserRole == 1 || (msgUserRole == 2 && groupMsg.SendUserID != req.UserID) {
						SendTipErrMsg(conn, "您没有权限撤回此消息")
						continue
					}
				}
				var content = "撤回了一条消息"
				content = "你" + content
				originMsg := groupMsg.Msg
				originMsg.WithdrawMsg = nil
				svcCtx.DB.Model(&groupMsg).Updates(group_models.GroupMsgModel{
					MsgPreview: "[撤回消息] - " + content,
					MsgType:    ctype.WithdrawMsgType,
					Msg: ctype.Msg{
						Type: ctype.WithdrawMsgType,
						WithdrawMsg: &ctype.WithdrawMsg{
							Content:   content,
							MsgID:     request.Msg.WithdrawMsg.MsgID,
							OriginMsg: &originMsg,
						},
					},
				})
			case ctype.ReplyMsgType:
				if request.Msg.ReplyMsg == nil || request.Msg.ReplyMsg.MsgID == 0 {
					SendTipErrMsg(conn, "回复消息的id必填")
					continue
				}
				var msgModel group_models.GroupMsgModel
				err = svcCtx.DB.Take(&msgModel, "group_id=? and id=?", request.GroupID, request.Msg.ReplyMsg.MsgID).Error
				if err != nil {
					SendTipErrMsg(conn, "回复消息不存在")
					continue
				}
				if msgModel.MsgType == ctype.WithdrawMsgType {
					SendTipErrMsg(conn, "消息已被撤回")
					continue
				}
				userBaseInfo, err5 := redis_service.GetUserBaseInfo(svcCtx.Redis, svcCtx.UserRpc, msgModel.SendUserID)
				if err5 != nil {
					logx.Error(err5)
					SendTipErrMsg(conn, err5.Error())
					continue
				}
				request.Msg.ReplyMsg.Msg = &msgModel.Msg
				request.Msg.ReplyMsg.UserID = msgModel.SendUserID
				request.Msg.ReplyMsg.UserNickName = userBaseInfo.NickName
				request.Msg.ReplyMsg.OriginMsgDate = msgModel.CreatedAt
				request.Msg.ReplyMsg.ReplyMsgPreview = msgModel.MsgPreview
			case ctype.QuoteMsgType:
				if request.Msg.QuoteMsg == nil || request.Msg.QuoteMsg.MsgID == 0 {
					SendTipErrMsg(conn, "引用消息的id必填")
					continue
				}
				var msgModel group_models.GroupMsgModel
				err = svcCtx.DB.Take(&msgModel, "group_id=? and id =? ", request.GroupID, request.Msg.QuoteMsg.MsgID).Error
				if err != nil {
					SendTipErrMsg(conn, "引用消息不存在")
					continue
				}
				if msgModel.MsgType == ctype.WithdrawMsgType {
					SendTipErrMsg(conn, "消息已被撤回")
					continue
				}
				userBaseInfo, err5 := redis_service.GetUserBaseInfo(svcCtx.Redis, svcCtx.UserRpc, msgModel.SendUserID)
				if err5 != nil {
					logx.Error(err5)
					SendTipErrMsg(conn, err5.Error())
					continue
				}
				request.Msg.QuoteMsg.Msg = &msgModel.Msg
				request.Msg.QuoteMsg.UserID = msgModel.SendUserID
				request.Msg.QuoteMsg.UserNickName = userBaseInfo.NickName
				request.Msg.QuoteMsg.OriginMsgDate = msgModel.CreatedAt
				request.Msg.QuoteMsg.QuoteMsgPreview = msgModel.MsgPreviewMethod()
			}
			msgID := insertMsg(svcCtx.DB, conn, member, request.Msg)
			SendGroupOnlineUserMsg(
				svcCtx.DB,
				member,
				request.Msg,
				msgID,
			)
		}
	}
}

func insertMsg(db *gorm.DB, conn *websocket.Conn, member group_models.GroupMemberModel, msg ctype.Msg) uint {
	switch msg.Type {
	case ctype.WithdrawMsgType:
		fmt.Println("撤回消息自己是不入库的")
		return 0
	}
	groupMsg := group_models.GroupMsgModel{
		GroupID:       member.GroupID,
		SendUserID:    member.UserID,
		GroupMemberID: member.ID,
		MsgType:       msg.Type,
		Msg:           msg,
	}
	groupMsg.MsgPreview = groupMsg.MsgPreviewMethod()
	err := db.Create(&groupMsg).Error
	if err != nil {
		logx.Error(err)
		SendTipErrMsg(conn, "消息入库失败")
		return 0
	}
	return groupMsg.ID
}
func SendGroupOnlineUserMsg(db *gorm.DB, member group_models.GroupMemberModel, msg ctype.Msg, msgID uint) {
	userOnlineIDList := getOnlineUserIDList()
	var groupMemberOnlineIDList []uint
	db.Model(group_models.GroupMemberModel{}).Where("group_id=?and user_id in ?", member.GroupID, userOnlineIDList).Select("user_id").Scan(&groupMemberOnlineIDList)
	var chatResponse = ChatResponse{
		GroupID:        member.GroupID,
		UserID:         member.UserID,
		Msg:            msg,
		ID:             msgID,
		MsgType:        msg.Type,
		CreatedAt:      time.Now(),
		MemberNickname: member.MemberNickname,
		MsgPreview:     msg.MsgPreview(),
	}
	wsInfo, ok := UserOnlineWsMap[member.UserID]
	if ok {
		chatResponse.UserNickname = wsInfo.UserINfo.NickName
		chatResponse.UserAvatar = wsInfo.UserINfo.Avatar
	}
	for _, u := range groupMemberOnlineIDList {
		wsUserInfo, ok2 := UserOnlineWsMap[u]
		if !ok2 {
			continue
		}
		chatResponse.IsMe = false
		if wsUserInfo.UserINfo.ID == member.UserID {
			chatResponse.IsMe = true
		}
		byteData, _ := json.Marshal(chatResponse)
		for _, w2 := range wsUserInfo.WsClientMap {
			w2.WriteMessage(websocket.TextMessage, byteData)
		}
	}
}
func getOnlineUserIDList() (userOnlineIDList []uint) {
	for u, _ := range UserOnlineWsMap {
		userOnlineIDList = append(userOnlineIDList, u)
	}
	return
}
func SendTipErrMsg(conn *websocket.Conn, msg string) {
	resp := ChatResponse{
		Msg: ctype.Msg{
			Type: ctype.TipMsgType,
			TipMsg: &ctype.TipMsg{
				Status:  "error",
				Content: msg,
			},
		},
		CreatedAt: time.Now(),
	}
	byteData, _ := json.Marshal(resp)
	conn.WriteMessage(websocket.TextMessage, byteData)
}
