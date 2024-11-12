package handler

import (
	"context"
	"encoding/json"
	"fim/common/models/ctype"
	"fim/common/response"
	"fim/common/service/redis_service"
	"fim/fim_chat/chat_api/internal/svc"
	"fim/fim_chat/chat_api/internal/types"
	"fim/fim_chat/chat_models"
	"fim/fim_file/file_rpc/types/file_rpc"
	"fim/fim_user/user_models"
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
	UserInfo    user_models.UserModel      //用户信息
	WsClientMap map[string]*websocket.Conn //这个用户管理所有ws连接
	currentConn *websocket.Conn            //当前连接
}

var UserOnlineWsMap = map[uint]*UserWsInfo{} //用户id和ws信息的映射
var VideoCallMap = map[string]time.Time{}    //音视频通话
type ChatResponse struct {
	ID         uint           `json:"id"`
	IsMe       bool           `json:"is_me"`
	RevUser    ctype.UserInfo `json:"revUser"`
	SendUser   ctype.UserInfo `json:"sendUser"`
	Msg        ctype.Msg      `json:"msg"`
	CreatedAt  time.Time      `json:"created_at"`
	MsgPreview string         `json:"msg_preview"`
}

// chatHandler 处理聊天请求的HTTP函数。
// 它负责升级HTTP连接到WebSocket连接，并管理用户的在线状态和聊天消息。
func chatHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 解析HTTP请求体中的聊天请求。
		var req types.ChatRequest
		if err := httpx.Parse(r, &req); err != nil {
			response.Response(r, w, nil, err)
			return
		}

		// 初始化WebSocket升级器，并允许所有来源的连接。
		var upGrader = websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}

		// 尝试升级HTTP连接到WebSocket连接。
		conn, err := upGrader.Upgrade(w, r, nil)
		if err != nil {
			logx.Error(err)
			response.Response(r, w, nil, err)
			return
		}

		// 获取并记录连接的远程地址，用于后续的连接管理。
		add := conn.RemoteAddr().String()
		defer func() {
			err := conn.Close()
			if err != nil {
				return
			}
			// 管理用户WebSocket连接的在线状态。
			userWsInfo, ok := UserOnlineWsMap[req.UserID]
			if ok {
				delete(userWsInfo.WsClientMap, add)
			}
			if userWsInfo != nil && len(userWsInfo.WsClientMap) == 0 {
				delete(UserOnlineWsMap, req.UserID)
				svcCtx.Redis.HDel("online", fmt.Sprintf("%d", req.UserID))
			}
		}()

		// 调用RPC服务获取用户信息。
		res, err := svcCtx.UserRpc.UserInfo(context.Background(), &user_rpc.UserInfoRequest{
			UserId: uint32(req.UserID),
		})
		if err != nil {
			logx.Error(err)
			response.Response(r, w, nil, err)
			return
		}

		// 解析用户信息。
		var userInfo user_models.UserModel
		err = json.Unmarshal(res.Data, &userInfo)
		if err != nil {
			logx.Error(err)
			response.Response(r, w, nil, err)
			return
		}

		// 管理用户的在线状态。
		userWsinfo, ok := UserOnlineWsMap[req.UserID]
		if !ok {
			userWsinfo = &UserWsInfo{
				UserInfo: userInfo,
				WsClientMap: map[string]*websocket.Conn{
					add: conn,
				},
				currentConn: conn,
			}
			UserOnlineWsMap[req.UserID] = userWsinfo
		}
		_, ok1 := userWsinfo.WsClientMap[add]
		if !ok1 {
			UserOnlineWsMap[req.UserID].WsClientMap[add] = conn
			UserOnlineWsMap[req.UserID].currentConn = conn
		}
		svcCtx.Redis.HSet("online", fmt.Sprintf("%d", req.UserID), req.UserID)

		// 获取用户的朋友列表。
		friendRes, err := svcCtx.UserRpc.FriendList(context.Background(), &user_rpc.FriendListRequest{
			User: uint32(req.UserID),
		})
		if err != nil {
			logx.Error(err)
			response.Response(r, w, nil, err)
			return
		}

		// 记录用户上线信息。
		logx.Infof("用户上线，%s 用户id；%d", userInfo.Nickname, req.UserID)

		// 通知用户的朋友关于其上线的消息。
		for _, info := range friendRes.FriendList {
			if uint(info.UserId) == req.UserID {
				continue
			}
			friend, ok := UserOnlineWsMap[uint(info.UserId)]
			if ok {
				text := fmt.Sprintf("好友%s上线了", UserOnlineWsMap[req.UserID].UserInfo.Nickname)
				if friend.UserInfo.UserConfModel.FriendOnline {
					resp := ChatResponse{
						Msg: ctype.Msg{
							Type: ctype.FriendOnlineMsgType,
							FriendOnlineMsg: &ctype.FriendOnlineMsg{
								NickName: userInfo.Nickname,
								Avatar:   userInfo.Avatar,
								Content:  text,
								FriendID: userInfo.ID,
							},
						},
						CreatedAt: time.Now(),
					}
					byteData, _ := json.Marshal(resp)
					sendMapMsg(friend.WsClientMap, byteData)
				}
			}
		}

		// 循环读取并处理WebSocket的消息。
		for {
			_, p, err1 := conn.ReadMessage()
			if err1 != nil {
				fmt.Println(err1)
				break
			}
			if userInfo.UserConfModel.CurtailChat {
				// 如果用户被限制聊天，则发送提示消息。
				SendTipErrMsg(conn, "你已被限制聊天，请联系客服")
				continue
			}
			var request Chatquest
			err2 := json.Unmarshal(p, &request)
			if err2 != nil {
				logx.Error(err2)
				SendTipErrMsg(conn, "参数解析失败")
				continue
			}
			if request.RevUserID != req.UserID {
				isFriendRes, err := svcCtx.UserRpc.IsFriend(context.Background(), &user_rpc.IsFriendRequest{
					User1: uint32(req.UserID),
					User2: uint32(request.RevUserID),
				})
				if err != nil {
					logx.Error(err)
					SendTipErrMsg(conn, "用户服务错误")
					continue
				}
				if !isFriendRes.IsFriend {
					SendTipErrMsg(conn, "你不是好友")
					continue
				}
			}
			if !(request.Msg.Type >= 1 && request.Msg.Type <= 14) {
				SendTipErrMsg(conn, "消息类型错误")
				continue
			}
			msgValidateErr := request.Msg.Validate()
			if msgValidateErr != nil {
				SendTipErrMsg(conn, msgValidateErr.Error())
				continue
			}
			switch request.Msg.Type {
			case ctype.TextMsgType:
			case ctype.FileMsgType:
				// 从文件源路径中获取文件名
				nameList := strings.Split(request.Msg.FileMsg.Src, "/")
				// 如果无法获取文件名，则提示用户并跳过当前循环
				if len(nameList) == 0 {
					SendTipErrMsg(conn, "请上传文件")
					continue
				}
				// 提取文件ID，即文件名部分
				fileID := nameList[len(nameList)-1]
				// 通过文件ID获取文件详细信息
				fileResponse, err3 := svcCtx.FileRpc.FileInfo(context.Background(), &file_rpc.FileInfoRequest{
					FileId: fileID,
				})
				// 如果获取文件信息失败，则记录错误并提示用户
				if err3 != nil {
					logx.Error(err3)
					SendTipErrMsg(conn, err3.Error())
					continue
				}
				// 更新文件消息的标题、大小和类型
				request.Msg.FileMsg.Title = fileResponse.FileName
				request.Msg.FileMsg.Size = fileResponse.FileSize
				request.Msg.FileMsg.Type = fileResponse.FileType

			// 撤回消息
			case ctype.WithdrawMsgType:
				// 检查撤回消息的ID是否为空
				if request.Msg.WithdrawMsg == nil {
					SendTipErrMsg(conn, "撤回消息id不能为空")
					continue
				}
				if request.Msg.WithdrawMsg.MsgID == 0 {
					SendTipErrMsg(conn, "撤回消息id不能为空")
					continue
				}
				// 只能撤回自己发送的消息，先找到消息的发送者
				var msgModel chat_models.ChatModel
				err = svcCtx.DB.Take(&msgModel, request.Msg.WithdrawMsg.MsgID).Error
				if err != nil {
					SendTipErrMsg(conn, "消息不存在")
					continue
				}
				// 撤回的消息不能再次撤回
				if msgModel.MsgType == ctype.WithdrawMsgType {
					SendTipErrMsg(conn, "撤回消息不能再撤回")
					continue
				}
				// 判断消息是否是当前用户发送的
				if msgModel.SendUserID != req.UserID {
					SendTipErrMsg(conn, "只能撤回自己发的消息")
					continue
				}
				// 只能撤回两分钟内的消息
				now := time.Now()
				subTime := now.Sub(msgModel.CreatedAt)
				if subTime >= time.Minute*2 {
					SendTipErrMsg(conn, "只能撤回2分钟内的消息")
					continue
				}

				// 构造撤回消息的内容
				var content = "撤回了一条消息"
				if userInfo.UserConfModel.RecallMessage != nil {
					content = "撤回了一条消息，" + *userInfo.UserConfModel.RecallMessage
				}

				// 保留原始消息内容，以便后续使用
				originMsg := msgModel.Msg
				// 清除原始消息的撤回消息字段，以准确反映其已被撤回的状态
				originMsg.WithdrawMsg = nil

				// 更新数据库中的消息记录
				// 解释为何这样更新：为了在聊天记录中明确标识该消息已被撤回，以及显示撤回的消息内容和类型
				svcCtx.DB.Model(&msgModel).Updates(chat_models.ChatModel{
					MsgPreview: "-[撤回消息]-" + content,  // 设置消息预览为撤回消息的标识和内容
					MsgType:    ctype.WithdrawMsgType, // 设置消息类型为撤回消息
					Msg: ctype.Msg{
						Type: ctype.WithdrawMsgType, // 详细记录消息类型为撤回
						WithdrawMsg: &ctype.WithdrawMsg{
							// 记录撤回消息的具体信息，包括内容、消息ID和原始消息
							Content:   content,
							MsgID:     request.Msg.WithdrawMsg.MsgID,
							OriginMsg: &originMsg,
						},
					},
				})

			// 处理回复消息类型的情况
			case ctype.ReplyMsgType:
				// 检查回复消息的ID是否有效
				if request.Msg.ReplyMsg == nil || request.Msg.ReplyMsg.MsgID == 0 {
					SendTipErrMsg(conn, "回复消息id不能为空")
					continue
				}
				// 从数据库中获取被回复的消息模型
				var msgModel chat_models.ChatModel
				err = svcCtx.DB.Take(&msgModel, request.Msg.ReplyMsg.MsgID).Error
				if err != nil {
					SendTipErrMsg(conn, "消息不存在")
					continue
				}
				// 检查消息是否已被撤回
				if msgModel.MsgType == ctype.WithdrawMsgType {
					SendTipErrMsg(conn, "消息已被撤回")
					continue
				}
				// 确保用户只能回复自己或对方的消息
				if !((msgModel.SendUserID == req.UserID && msgModel.RevUserID == request.RevUserID) ||
					(msgModel.SendUserID == request.RevUserID && msgModel.RevUserID == req.UserID)) {
					SendTipErrMsg(conn, "只能回复自己的消息或对方的消息")
					continue
				}
				// 获取发送用户的基本信息
				userBaseInfo, err5 := redis_service.GetUserBaseInfo(svcCtx.Redis, svcCtx.UserRpc, msgModel.SendUserID)
				if err5 != nil {
					logx.Error(err5)
					SendTipErrMsg(conn, err5.Error())
					continue
				}
				// 更新回复消息的信息，包括消息内容、发送用户ID、昵称和原始发送时间
				request.Msg.ReplyMsg.Msg = &msgModel.Msg
				request.Msg.ReplyMsg.UserID = msgModel.SendUserID
				request.Msg.ReplyMsg.UserNickName = userBaseInfo.NickName
				request.Msg.ReplyMsg.OriginMsgDate = msgModel.CreatedAt
				// 生成消息预览
				request.Msg.ReplyMsg.ReplyMsgPreview = msgModel.MsgPreviewMethod()

			case ctype.QuoteMsgType:
				if request.Msg.QuoteMsg == nil || request.Msg.QuoteMsg.MsgID == 0 {
					SendTipErrMsg(conn, "请选择要引用的消息")
					continue
				}
				var msgModel chat_models.ChatModel
				err = svcCtx.DB.Take(&msgModel, request.Msg.QuoteMsg.MsgID).Error

				if err != nil {
					SendTipErrMsg(conn, "消息不存在")
					continue
				}
				if msgModel.MsgType == ctype.WithdrawMsgType {
					SendTipErrMsg(conn, "消息已被撤回")
					continue
				}

				if !((msgModel.SendUserID == req.UserID && msgModel.RevUserID == request.RevUserID) ||
					(msgModel.SendUserID == request.RevUserID && msgModel.RevUserID == req.UserID)) {
					SendTipErrMsg(conn, "只能引用自己的消息或对方的消息")
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
				request.Msg.QuoteMsg.QuoteMsgPreview = msgModel.MsgPreviewMethod()

			case ctype.VideoCallMsgType:
				data := request.Msg.VideoCallMsg
				_, ok2 := UserOnlineWsMap[request.RevUserID]
				if !ok2 {
					SendTipErrMsg(conn, "对方不在线")
					continue
				}
				key := fmt.Sprintf("video_call_%d_%d", req.UserID, request.RevUserID)
				switch data.Flag {
				case 0:
					//	给自己页面展示等待对方接听的弹框
					err := conn.WriteJSON(ChatResponse{
						Msg: ctype.Msg{
							Type: ctype.VideoCallMsgType,
							VideoCallMsg: &ctype.VideoCallMsg{
								Flag: 1,
							},
						},
					})
					if err != nil {
						return
					}

					sendRevUserMsg(request.RevUserID, req.UserID, ctype.Msg{
						Type: ctype.VideoCallMsgType,
						VideoCallMsg: &ctype.VideoCallMsg{
							Flag: 2,
						},
					})
				case 1: //自己挂断
					sendRevUserMsg(request.RevUserID, req.UserID, ctype.Msg{
						Type: ctype.VideoCallMsgType,
						VideoCallMsg: &ctype.VideoCallMsg{
							Flag: 3,
							Msg:  "发起者已挂断",
						},
					})
				case 2: //对方挂断
					sendRevUserMsg(request.RevUserID, req.UserID, ctype.Msg{
						Type: ctype.VideoCallMsgType,
						VideoCallMsg: &ctype.VideoCallMsg{
							Flag: 4,
							Msg:  "对方拒绝接听",
						},
					})
				case 3: //对方接听
					sendRevUserMsg(request.RevUserID, req.UserID, ctype.Msg{
						Type: ctype.VideoCallMsgType,
						VideoCallMsg: &ctype.VideoCallMsg{
							Flag: 5,
							Type: "create_offer",
						},
					})

				case 4: // 正常挂断
					// 根据通话双方的ID获取通话开始时间
					startTime, ok3 := VideoCallMap[key]
					// 定义发送方和接收方的用户ID
					var sendUserID = req.UserID
					var revUserID = request.RevUserID
					// 打印日志，记录当前处理的通话双方ID
					fmt.Println("key1", key, sendUserID, revUserID)
					// 定义通话结束原因变量
					var endReason int8
					// 如果当前通话双方的开始时间不存在，则尝试交换双方ID后再次查找
					if !ok3 {
						key = fmt.Sprintf("%d %d", sendUserID, revUserID)
						_startTime, ok4 := VideoCallMap[key]
						// 如果交换ID后仍找不到开始时间，则提示错误并继续下一轮循环
						if !ok4 {
							SendTipErrMsg(conn, "通话起始时间错误")
							continue
						}
						// 交换通话双方ID
						sendUserID = request.RevUserID
						revUserID = userInfo.ID
						// 设置通话结束原因为1，表示正常挂断
						endReason = 1
						// 更新开始时间
						startTime = _startTime
					}
					// 打印日志，记录调整后的通话双方ID
					fmt.Println("key2", key, sendUserID, revUserID)
					// 计算通话时长
					subTime := time.Now().Sub(startTime)
					// 打印日志，记录当前通话的时长
					fmt.Printf("用户正常挂断，视频通话时长为%s\n", subTime)
					// 更新请求中的通话开始和结束时间以及结束原因
					request.Msg.VideoCallMsg.StartTime = startTime
					request.Msg.VideoCallMsg.EndTime = time.Now()
					request.Msg.VideoCallMsg.EndReason = endReason
					// 发送挂断消息给通话双方
					sendRevUserMsg(sendUserID, revUserID, ctype.Msg{
						Type: ctype.VideoCallMsgType,
						VideoCallMsg: &ctype.VideoCallMsg{
							Flag: 6,
							Msg:  "正常挂断",
						},
					})
					// 将通话记录插入数据库，并获取消息ID
					msgID := InsertMsgByChat(svcCtx.DB, revUserID, sendUserID, request.Msg)
					// 根据用户ID发送消息
					SendMsgByUser(svcCtx, revUserID, sendUserID, request.Msg, msgID)
					// 从通话映射中删除当前通话记录
					delete(VideoCallMap, key)

				case 5:
					//对方挂断
					key = fmt.Sprintf("%d_%d", request.RevUserID, userInfo.ID)
					startTime, ok3 := VideoCallMap[key]
					if !ok3 {
						SendTipErrMsg(conn, "通话起始时间错误")
						continue
					}
					subTime := time.Now().Sub(startTime)
					fmt.Printf("用户挂断，视频通话时长为%s\n", subTime)
				}
				// 根据数据类型处理不同的WebRTC信令消息
				switch data.Type {
				case "offer": // 处理发起通话请求
					// 向接收用户发送通话请求消息
					sendRevUserMsg(request.RevUserID, req.UserID, ctype.Msg{
						Type: ctype.VideoCallMsgType,
						VideoCallMsg: &ctype.VideoCallMsg{
							Type: "offer",
							Data: data.Data,
						},
					})
					// 更新通话请求时间
					VideoCallMap[key] = time.Now()
					fmt.Println("offer", key)
				case "answer": // 处理通话应答
					// 向接收用户发送通话应答消息
					sendRevUserMsg(request.RevUserID, req.UserID, ctype.Msg{
						Type: ctype.WithdrawMsgType,
						VideoCallMsg: &ctype.VideoCallMsg{
							Type: "answer",
							Data: data.Data,
						},
					})
				case "offer_ice": // 处理通话请求的ICE候选
					// 向接收用户发送ICE候选消息
					sendRevUserMsg(request.RevUserID, req.UserID, ctype.Msg{
						Type: ctype.VideoCallMsgType,
						VideoCallMsg: &ctype.VideoCallMsg{
							Type: "offer_ice",
							Data: data.Data,
						},
					})
				case "answer_ice": // 处理通话应答的ICE候选
					// 向接收用户发送ICE候选消息
					sendRevUserMsg(request.RevUserID, req.UserID, ctype.Msg{
						Type: ctype.VideoCallMsgType,
						VideoCallMsg: &ctype.VideoCallMsg{
							Type: "answer_ice",
							Data: data.Data,
						},
					})
				}
				// 继续处理其他消息
				continue
			}

			msgID := InsertMsgByChat(svcCtx.DB, request.RevUserID, req.UserID, request.Msg)
			SendMsgByUser(svcCtx, request.RevUserID, req.UserID, request.Msg, msgID)
		}
	}
}

type Chatquest struct {
	RevUserID uint      `json:"revUserID"`
	Msg       ctype.Msg `json:"msg"`
}

// SendMsgByUser 根据用户ID发送消息。
// svcCtx: 服务上下文，用于访问服务环境。
// revUserID: 接收消息的用户ID。
// sendUserID: 发送消息的用户ID。
// msg: 要发送的消息内容。
// msgID: 消息的唯一ID。
func SendMsgByUser(svcCtx *svc.ServiceContext, revUserID uint, sendUserID uint, msg ctype.Msg, msgID uint) {
	// 从在线用户WebSocket映射中获取接收者和发送者的连接。
	revUser, ok1 := UserOnlineWsMap[revUserID]
	sendUser, ok2 := UserOnlineWsMap[sendUserID]

	// 构建聊天响应对象。
	resp := ChatResponse{
		ID:         msgID,
		Msg:        msg,
		MsgPreview: msg.MsgPreview(), // 提取消息预览。
		CreatedAt:  time.Now(),       // 记录当前时间作为创建时间。
	}

	// 当接收者、发送者都在在线映射中且不是自言自语时，构建详细的用户信息并发送消息。
	if ok1 && ok2 && sendUserID == revUserID {
		resp.RevUser = ctype.UserInfo{
			ID:       revUserID,
			NickName: revUser.UserInfo.Nickname,
			Avatar:   revUser.UserInfo.Avatar,
		}
		resp.SendUser = ctype.UserInfo{
			ID:       sendUserID,
			NickName: sendUser.UserInfo.Nickname,
			Avatar:   sendUser.UserInfo.Avatar,
		}
		resp.IsMe = true // 标记为是自己发送的消息。

		// 将响应对象序列化为字节切片并发送给接收者的WebSocket客户端。
		byteData, _ := json.Marshal(resp)
		sendWsMapMsg(revUser.WsClientMap, byteData)
		return
	}
	// 处理接收者不在线的情况。
	if !ok1 {
		userBaseInfo, err := redis_service.GetUserBaseInfo(svcCtx.Redis, svcCtx.UserRpc, revUserID)
		if err != nil {
			logx.Error(err)
			return
		}
		resp.RevUser = ctype.UserInfo{
			ID:       revUserID,
			NickName: userBaseInfo.NickName,
			Avatar:   userBaseInfo.Avatar,
		}
	} else {
		resp.RevUser = ctype.UserInfo{
			ID:       revUserID,
			NickName: revUser.UserInfo.Nickname,
			Avatar:   revUser.UserInfo.Avatar,
		}
	}
	// 设置发送用户信息。
	resp.SendUser = ctype.UserInfo{
		ID:       sendUserID,
		NickName: sendUser.UserInfo.Nickname,
		Avatar:   sendUser.UserInfo.Avatar,
	}
	resp.IsMe = true
	byteData, _ := json.Marshal(resp)
	sendWsMapMsg(sendUser.WsClientMap, byteData)
	// 如果接收者在线，则发送消息给接收者。
	if ok1 {
		resp.IsMe = false
		byteData, _ := json.Marshal(resp)
		sendWsMapMsg(revUser.WsClientMap, byteData)
	}
}

// sendWsMapMsg 向所有在wsMap中的WebSocket连接发送数据。
//
// 参数:
// wsMap - 一个映射，其中包含所有需要发送数据的WebSocket连接，键为连接的标识符，值为连接对象。
// byteData - 要发送的数据，类型为字节切片。
//
// 该函数遍历wsMap中的所有连接，并向每个连接发送byteData数据。发送的数据类型为文本消息。
func sendWsMapMsg(wsMap map[string]*websocket.Conn, byteData []byte) {
	for _, conn := range wsMap {
		err := conn.WriteMessage(websocket.TextMessage, byteData)
		if err != nil {
			return
		}
	}
}

// InsertMsgByChat 根据聊天内容插入消息到数据库。
// 参数:
// db: Gorm数据库连接实例，用于执行数据库操作。
// revUserID: 接收用户的ID。
// sendUserID: 发送用户的ID。
// msg: 待插入的消息对象。
// 返回值:
// msgID: 插入消息的ID，若操作失败则返回0。
func InsertMsgByChat(db *gorm.DB, revUserID uint, sendUserID uint, msg ctype.Msg) (msgID uint) {
	// 处理撤回消息的特殊情况，撤回消息通常不需要存入数据库。
	if msg.Type == ctype.WithdrawMsgType {
		fmt.Println("撤回消息自己是不入库的")
		return
	}

	// 初始化ChatModel对象，准备将消息数据存入数据库。
	chatModel := chat_models.ChatModel{
		SendUserID: sendUserID,
		RevUserID:  revUserID,
		MsgType:    msg.Type,
		Msg:        msg,
	}

	// 生成消息预览并赋值给ChatModel。
	chatModel.MsgPreview = chatModel.MsgPreviewMethod()

	// 尝试将消息数据插入数据库。
	err := db.Create(&chatModel).Error
	if err != nil {
		// 记录错误日志，并尝试通知发送用户消息入库失败。
		logx.Error(err)
		sendUser, ok := UserOnlineWsMap[sendUserID]
		if !ok {
			return
		}
		SendTipErrMsg(sendUser.currentConn, "消息入库失败")
	}

	// 返回成功插入消息的ID。
	return chatModel.ID
}

// sendMapMsg 向websocket连接映射中的所有连接发送字节数据。
// wsMap: 连接映射，键为连接标识，值为websocket连接对象。
// byteData: 需要发送的字节数据。
func sendMapMsg(wsMap map[string]*websocket.Conn, byteData []byte) {
	// 遍历映射中的所有连接
	for _, conn := range wsMap {
		// 向每个连接发送字节数据
		err := conn.WriteMessage(websocket.TextMessage, byteData)
		if err != nil {
			return
		}
	}
}

// sendRevUserMsg 向指定接收者发送消息
// 参数:
// revUserID - 接收消息的用户ID
// sendUserID - 发送消息的用户ID
// msg - 要发送的消息对象
func sendRevUserMsg(revUserID uint, sendUserID uint, msg ctype.Msg) {
	// 尝试从在线用户映射中获取接收者的信息
	userRes, ok := UserOnlineWsMap[revUserID]
	// 如果接收者不在线，则直接返回
	if !ok {
		return
	}

	// 尝试从在线用户映射中获取发送者的信息
	sendUser, ok1 := UserOnlineWsMap[sendUserID]
	var sendUserInfo ctype.UserInfo
	// 如果发送者在线，则构建发送者的用户信息
	if ok1 {
		sendUserInfo = ctype.UserInfo{
			ID:       sendUser.UserInfo.ID,
			NickName: sendUser.UserInfo.Nickname,
			Avatar:   sendUser.UserInfo.Avatar,
		}
	}

	// 遍历接收者的WebSocket连接并发送消息
	for _, conn := range userRes.WsClientMap {
		// 构建并发送聊天响应
		err := conn.WriteJSON(ChatResponse{
			SendUser: sendUserInfo,
			RevUser: ctype.UserInfo{
				ID:       userRes.UserInfo.ID,
				NickName: userRes.UserInfo.Nickname,
				Avatar:   userRes.UserInfo.Avatar,
			},
			MsgPreview: msg.MsgPreview(),
			Msg:        msg,
			CreatedAt:  time.Now(),
		})
		if err != nil {
			return
		}
		// 发送完一条消息后即退出循环，这里可能需要解释为什么这样做
		break
	}
}

// SendTipErrMsg 向指定的websocket连接发送提示错误消息。
// conn: 需要发送消息的websocket连接对象。
// msg: 错误提示内容。
func SendTipErrMsg(conn *websocket.Conn, msg string) {
	// 构建响应消息，包含错误提示信息
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
	// 将响应消息序列化为JSON字节数据
	byteData, _ := json.Marshal(resp)
	// 向连接发送序列化后的错误提示消息
	err := conn.WriteMessage(websocket.TextMessage, byteData)
	if err != nil {
		return
	}
}
