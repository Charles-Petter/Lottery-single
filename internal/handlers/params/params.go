package params

import (
	"lottery_single/internal/service"
	"time"
)

// PrizeListRequest 处理请求和响应的实体
type PrizeListRequest struct {
}

type PrizeListResponse struct {
}

type LoginReq struct {
	UserName string `json:"user_name"`
	PassWord string `json:"pass_word"`
}

type LoginRsp struct {
	UserID int
	Token  string
}

// RegisterReq 定义注册请求所需的参数
type RegisterReq struct {
	UserName string `form:"user_name" json:"user_name" binding:"required"`
	PassWord string `form:"pass_word" json:"pass_word" binding:"required"`
	Email    string `form:"email" json:"email" binding:"required,email"`
	Mobile   string `form:"mobile" json:"mobile" binding:"required"`
	RealName string `form:"real_name" json:"real_name" binding:"required"`
	Age      int    `form:"age" json:"age" binding:"required,min=0"`
	Gender   string `form:"gender" json:"gender" binding:"required,oneof=male female"`
}
type LotteryReq struct {
	UserID uint   `json:"user_id"`
	Token  string `json:"token"`
	IP     string `json:"ip"`
}

type PrizeAddRequest struct {
	PrizeInfo service.ViewPrize
}

type BlackIp struct {
	ID         uint      `gorm:"primaryKey"`
	Ip         string    `gorm:"unique"`
	BlackTime  time.Time `gorm:"default:null"`
	SysCreated time.Time `gorm:"autoCreateTime"`
	SysUpdated time.Time `gorm:"autoUpdateTime"`
}

// CouponListRequest 请求优惠券列表的参数
type CouponListRequest struct {
	PrizeID uint `json:"prize_id"`
}

// CouponListResponse 优惠券列表的响应
type CouponListResponse struct {
	Coupons        []*service.ViewCouponInfo `json:"coupons"`
	DBCouponNum    int64                     `json:"db_coupon_num"`
	CacheCouponNum int64                     `json:"cache_coupon_num"`
}
