// Code generated by goctl. DO NOT EDIT.
package types

type AddFriendRequest struct {
	UserID               uint                  `header:"user_id"`
	FriendID             uint                  `json:"friend_id"`
	Verify               string                `json:"verify",optional` // 验证信息
	VerificationQuestion *VerificationQuestion `json:"verification_question,optional"`
}

type AddFriendResponse struct {
}

type DeleteFriendRequest struct {
	UserID   uint `header:"user_id"`
	FriendID uint `json:"friend_id"`
}

type DeleteFriendResponse struct {
}

type FriendInfoRequest struct {
	UserID   uint `header:"user_id"`
	Role     int8 `header:"Role"`
	FriendID uint `form:"friend_id"`
}

type FriendInfoResponse struct {
	UserID   uint   `json:"user_id"`
	Nickname string `json:"nickname"`
	Abstract string `json:"abstract"`
	Avatar   string `json:"avatar"`
	Notice   string `json:"notice"`
	IsOnline bool   `json:"isOnline"` // 是否在线
}

type FriendListRequest struct {
	UserID uint `header:"user_id"`
	Role   int8 `header:"Role"`
	Page   int  `form:"page,optional"`
	Limit  int  `form:"limit,optional"`
}

type FriendListResponse struct {
	List  []FriendInfoResponse `json:"list"`
	Count int                  `json:"count"`
}

type FriendNoticeUpdateRequest struct {
	UserID   uint   `header:"user_id"`
	FriendID uint   `json:"friend_id"`
	Notice   string `json:"notice"` // 备注
}

type FriendNoticeUpdateResponse struct {
}

type FriendValidInfo struct {
	UserID               uint                  `json:"user_id"`
	Nickname             string                `json:"nickname"`
	AdditionalMessages   string                `json:"additional_messages"`   // 附加信息
	Avatar               string                `json:"avatar"`                // 头像
	VerificationQuestion *VerificationQuestion `json:"verification_question"` // 验证问题 为3和4的时候需要
	Status               int8                  `json:"status"`                // 验证状态  0 未操作 1 同意 2 拒绝 3 忽略
	SendStatus           int8                  `json:"send_status"`           // 发送状态 0 未操作 1 同意 2 拒绝 3 忽略
	RevStatus            int8                  `json:"rev_status"`            // 接收状态 0 未操作 1 同意 2 拒绝 3 忽略
	Verification         int8                  `json:"verification"`          // 验证状态状态
	ID                   uint                  `json:"id"`                    // 验证id
	Flag                 string                `json:"flag"`                  // 标识 send or rev
	CreatedAt            string                `json:"created_at"`            // 创建时间
}

type FriendValidRequest struct {
	UserID uint `header:"user_id"`
	Page   int  `form:"page,optional"`
	Limit  int  `form:"limit,optional"`
}

type FriendValidResponse struct {
	List  []FriendValidInfo `json:"list"`
	Count int64             `json:"count"`
}

type FriendValidStatusRequest struct {
	UserID   uint `header:"user_id"`
	VerifyID uint `json:"verify_id"`
	Status   int8 `json:"status"` // 验证状态
}

type FriendValidStatusResponse struct {
}

type SearchInfo struct {
	UserID   uint   `json:"user_id"`
	Nickname string `json:"nickname"`
	Abstract string `json:"abstract"`
	Avatar   string `json:"avatar"`
	IsFriend bool   `json:"isFriend"` // 是否是好友
}

type SearchRequest struct {
	UserID uint   `header:"user_id "`
	Key    string `form:"key,optional"`    // 用户id和昵称
	Online bool   `form:"online,optional"` // 搜索在线的用户
	Page   int    `form:"page,optional"`
	Limit  int    `form:"limit,optional"`
}

type SearchResponse struct {
	List  []SearchInfo `json:"list"`
	Count int64        `json:"count"`
}

type UserInfoRequest struct {
	UserID uint `header:"user_id"`
	Role   int8 `header:"Role"`
}

type UserInfoResponse struct {
	UserID               uint                  `json:"user_id"`
	Nickname             string                `json:"nickname"`
	Abstract             string                `json:"abstract"`
	Avatar               string                `json:"avatar"`
	RecallMessage        *string               `json:"recall_message"`
	FriendOnline         bool                  `json:"friend_online"`
	Sound                bool                  `json:"sound"`
	SecureLink           bool                  `json:"secure_link"`
	SavePwd              bool                  `json:"save_pwd"`
	SearchUser           int8                  `json:"search_user"`
	Verification         int8                  `json:"verification"`
	VerificationQuestion *VerificationQuestion `json:"verification_question"`
}

type UserInfoUpdateRequest struct {
	UserID               uint                  `header:"user_id"`
	Nickname             *string               `json:"nickname,optional" user:"nickname"`
	Abstract             *string               `json:"abstract,optional" user:"abstract"`
	Avatar               *string               `json:"avatar,optional" user:"avatar"`
	RecallMessage        *string               `json:"recallMessage,optional" user_conf:"recall_message"`
	FriendOnline         *bool                 `json:"friendOnline,optional" user_conf:"friend_online"`
	Sound                *bool                 `json:"sound,optional" user_conf:"sound"`
	SecureLink           *bool                 `json:"secureLink,optional" user_conf:"secure_link"`
	SavePwd              *bool                 `json:"savePwd,optional" user_conf:"save_pwd"`
	SearchUser           *int8                 `json:"searchUser,optional" user_conf:"search_user"`
	Verification         *int8                 `json:"verification,optional" user_conf:"verification"`
	VerificationQuestion *VerificationQuestion `json:"verificationQuestion,optional" user_conf:"verification_question"`
}

type UserInfoUpdateResponse struct {
}

type UserVaildRequest struct {
	UserID   uint `header:"user_id"`
	FriendID uint `json:"friend_id"`
}

type UserVaildResponse struct {
	Verification         int8                 `json:"verification"`          // 验证状态
	VerificationQuestion VerificationQuestion `json:"verification_question"` // 验证问题
}

type VerificationQuestion struct {
	Problem1 *string `json:"problem1,optional"user_conf:"problem1"`
	Problem2 *string `json:"problem2,optional"user_conf:"problem2"`
	Problem3 *string `json:"problem3,optional"user_conf:"problem3"`
	Answer1  *string `json:"answer1,optional"user_conf:"answer1"`
	Answer2  *string `json:"answer2,optional"user_conf:"answer2"`
	Answer3  *string `json:"answer3,optional"user_conf:"answer3"`
}
