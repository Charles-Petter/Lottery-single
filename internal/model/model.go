package model

import "time"

// Prize 奖品表
type Prize struct {
	Id           uint       `gorm:"column:id;type:int(10) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	Title        string     `gorm:"column:title;type:varchar(255);comment:奖品名称;NOT NULL" json:"title"`
	PrizeNum     int        `gorm:"column:prize_num;type:int(11);default:-1;comment:奖品数量，0 无限量，>0限量，<0无奖品;NOT NULL" json:"prize_num"`
	LeftNum      int        `gorm:"column:left_num;type:int(11);default:0;comment:剩余数量;NOT NULL" json:"left_num"`
	PrizeCode    string     `gorm:"column:prize_code;type:varchar(50);comment:0-9999表示100%，0-0表示万分之一的中奖概率;NOT NULL" json:"prize_code"`
	PrizeTime    uint       `gorm:"column:prize_time;type:int(10) unsigned;default:0;comment:发奖周期，多少天，以天为单位;NOT NULL" json:"prize_time"`
	Img          string     `gorm:"column:img;type:varchar(255);comment:奖品图片;NOT NULL" json:"img"`
	DisplayOrder uint       `gorm:"column:display_order;type:int(10) unsigned;default:0;comment:位置序号，小的排在前面;NOT NULL" json:"display_order"`
	PrizeType    uint       `gorm:"column:prize_type;type:int(10) unsigned;default:0;comment:奖品类型，0 虚拟币，1 虚拟券，2 实物-小奖，3 实物-大奖;NOT NULL" json:"prize_type"`
	PrizeProfile string     `gorm:"column:prize_profile;type:varchar(255);comment:奖品扩展数据，如：虚拟币数量;NOT NULL" json:"prize_profile"`
	BeginTime    time.Time  `gorm:"column:begin_time;type:datetime;default:1000-01-01 00:00:00;comment:奖品有效周期：开始时间;NOT NULL" json:"begin_time"`
	EndTime      time.Time  `gorm:"column:end_time;type:datetime;default:1000-01-01 00:00:00;comment:奖品有效周期：结束时间;NOT NULL" json:"end_time"`
	PrizePlan    string     `gorm:"column:prize_plan;type:mediumtext;comment:发奖计划，[[时间1,数量1],[时间2,数量2]]" json:"prize_plan"`
	PrizeBegin   time.Time  `gorm:"column:prize_begin;type:int(11);default:1000-01-01 00:00:00;comment:发奖计划周期的开始;NOT NULL" json:"prize_begin"`
	PrizeEnd     time.Time  `gorm:"column:prize_end;type:int(11);default:1000-01-01 00:00:00;comment:发奖计划周期的结束;NOT NULL" json:"prize_end"`
	SysStatus    uint       `gorm:"column:sys_status;type:smallint(5) unsigned;default:0;comment:状态，0 正常，1 删除;NOT NULL" json:"sys_status"`
	SysCreated   *time.Time `gorm:"autoCreateTime:datetime;column:sys_created;type:datetime;default null;comment:创建时间;NOT NULL" json:"sys_created"`
	SysUpdated   *time.Time `gorm:"autoUpdateTime:datetime;column:sys_updated;type:datetime;default null;comment:修改时间;NOT NULL" json:"sys_updated"`
	SysIp        string     `gorm:"column:sys_ip;type:varchar(50);comment:操作人IP;NOT NULL" json:"sys_ip"`
}

func (p *Prize) TableName() string {
	return "t_prize"
}

// User 用户表
type User struct {
	Id        uint   `gorm:"column:id;type:int(10) unsigned;AUTO_INCREMENT;NOT NULL" json:"id"`
	UserName  string `gorm:"column:user_name;type:varchar(255);comment:用户名称;NOT NULL" json:"user_name"`
	Password  string `gorm:"column:pass_word;type:varchar(255);comment:用户密码;NOT NULL" json:"pass_word"`
	Signature string `gorm:"column:signature;type:varchar(1024);comment:签名" json:"signature"`
	Email     string `json:"email"`
	Mobile    string `json:"mobile"`
	RealName  string `json:"real_name"`
	Age       int    `json:"age"`
	Gender    string `json:"gender"`
}

func (u *User) TableName() string {
	return "t_user"
}

// Coupon 优惠券表
type Coupon struct {
	Id         uint       `gorm:"column:id;type:int(10) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	PrizeId    uint       `gorm:"column:prize_id;type:int(10) unsigned;default:0;comment:奖品ID，关联lt_prize表;NOT NULL" json:"prize_id"`
	Code       string     `gorm:"column:code;type:varchar(255);comment:虚拟券编码;NOT NULL" json:"code"`
	SysCreated *time.Time `gorm:"autoCreateTime;column:sys_created;type:datetime;default null;comment:创建时间;NOT NULL" json:"sys_created"`
	SysUpdated *time.Time `gorm:"autoUpdateTime;column:sys_updated;type:datetime;default null;comment:更新时间;NOT NULL" json:"sys_updated"`
	SysStatus  uint       `gorm:"column:sys_status;type:smallint(5) unsigned;default:0;comment:状态，1正常，2作废，2已发放;NOT NULL" json:"sys_status"`
}

func (c *Coupon) TableName() string {
	return "t_coupon"
}

// Result 抽奖记录表
type Result struct {
	Id         uint       `gorm:"column:id;type:int(10) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	PrizeId    uint       `gorm:"column:prize_id;type:int(10) unsigned;default:0;comment:奖品ID，关联lt_prize表;NOT NULL" json:"prize_id"`
	PrizeName  string     `gorm:"column:prize_name;type:varchar(255);comment:奖品名称;NOT NULL" json:"prize_name"`
	PrizeType  uint       `gorm:"column:prize_type;type:int(10) unsigned;default:0;comment:奖品类型，同lt_prize. gtype;NOT NULL" json:"prize_type"`
	UserId     uint       `gorm:"column:user_id;type:int(10) unsigned;default:0;comment:用户ID;NOT NULL" json:"user_id"`
	UserName   string     `gorm:"column:user_name;type:varchar(50);comment:用户名;NOT NULL" json:"user_name"`
	PrizeCode  uint       `gorm:"column:prize_code;type:int(10) unsigned;default:0;comment:抽奖编号（4位的随机数）;NOT NULL" json:"prize_code"`
	PrizeData  string     `gorm:"column:prize_data;type:varchar(255);comment:获奖信息;NOT NULL" json:"prize_data"`
	SysCreated *time.Time `gorm:"autoCreateTime;column:sys_created;type:datetime;default null;comment:创建时间;NOT NULL" json:"sys_created"`
	SysIp      string     `gorm:"column:sys_ip;type:varchar(50);comment:用户抽奖的IP;NOT NULL" json:"sys_ip"`
	SysStatus  uint       `gorm:"column:sys_status;type:smallint(5) unsigned;default:0;comment:状态，0 正常，1删除，2作弊;NOT NULL" json:"sys_status"`
}

func (r *Result) TableName() string {
	return "t_result"
}

// BlackUser 用户黑明单表
type BlackUser struct {
	Id         uint       `gorm:"column:id;type:int(10) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	UserId     uint       `gorm:"column:user_id;type:int(10) unsigned;default:0;comment:用户ID;NOT NULL" json:"user_id"`
	UserName   string     `gorm:"column:user_name;type:varchar(50);comment:用户名;NOT NULL" json:"user_name"`
	BlackTime  time.Time  `gorm:"column:black_time;type:datetime;default:1000-01-01 00:00:00;comment:黑名单限制到期时间;NOT NULL" json:"black_time"`
	RealName   string     `gorm:"column:real_name;type:varchar(50);comment:真是姓名;NOT NULL" json:"real_name"`
	Mobile     string     `gorm:"column:mobile;type:varchar(50);comment:手机号;NOT NULL" json:"mobile"`
	Address    string     `gorm:"column:address;type:varchar(255);comment:联系地址;NOT NULL" json:"address"`
	SysCreated *time.Time `gorm:"autoCreateTime;column:sys_created;type:datetime;default null;comment:创建时间;NOT NULL" json:"sys_created"`
	SysUpdated *time.Time `gorm:"autoUpdateTime;column:sys_updated;type:datetime;default null;comment:修改时间;NOT NULL" json:"sys_updated"`
	SysIp      string     `gorm:"column:sys_ip;type:varchar(50);comment:IP地址;NOT NULL" json:"sys_ip"`
}

func (m *BlackUser) TableName() string {
	return "t_black_user"
}

// BlackIp ip黑明单表
type BlackIp struct {
	Id         uint       `gorm:"column:id;type:int(10) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	Ip         string     `gorm:"column:ip;type:varchar(50);comment:IP地址;NOT NULL" json:"ip"`
	BlackTime  time.Time  `gorm:"column:black_time;type:datetime;default:1000-01-01 00:00:00;comment:黑名单限制到期时间;NOT NULL" json:"black_time"`
	SysCreated *time.Time `gorm:"autoCreateTime;column:sys_created;type:datetime;default null;comment:创建时间;NOT NULL" json:"sys_created"`
	SysUpdated *time.Time `gorm:"autoUpdateTime;column:sys_updated;type:datetime;default null;comment:修改时间;NOT NULL" json:"sys_updated"`
}

func (m *BlackIp) TableName() string {
	return "t_black_ip"
}

// LotteryTimes 用户每日抽奖次数表
type LotteryTimes struct {
	Id         uint       `gorm:"column:id;type:int(10) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	UserId     uint       `gorm:"column:user_id;type:int(10) unsigned;default:0;comment:用户ID;NOT NULL" json:"user_id"`
	Day        uint       `gorm:"column:day;type:int(10) unsigned;default:0;comment:日期，如：20220625;NOT NULL" json:"day"`
	Num        uint       `gorm:"column:num;type:int(10) unsigned;default:0;comment:次数;NOT NULL" json:"num"`
	SysCreated *time.Time `gorm:"autoCreateTime;column:sys_created;type:datetime;default null;comment:创建时间;NOT NULL" json:"sys_created"`
	SysUpdated *time.Time `gorm:"autoUpdateTime;column:sys_updated;type:datetime;default null;comment:修改时间;NOT NULL" json:"sys_updated"`
}

func (l *LotteryTimes) TableName() string {
	return "t_lottery_times"
}

type Teacher struct {
	Id          int        `gorm:"primaryKey;autoIncrement;comment:主键id"` //所谓蛇形复数
	Tno         int        `gorm:"default:0"`
	Name        string     `gorm:"type:varchar(10);not null"`
	Pwd         string     `gorm:"type:varchar(100);not null"`
	Tel         string     `gorm:"type:char(11);column:my_name"`
	Birth       *time.Time //它的零值（默认值）将是time.Time{}，而不是 nil，因为 time.Time 是值类型，它的默认值是其零值。如果你想要在这个字段中存储 NULL 值，就需要使用 *time.Time 类型，并将其设置为 nil。
	Remark      string     `gorm:"type:varchar(255);"`
	CreatTime   *time.Time `gorm:"autoCreateTime;default null"`
	DeletedTime *time.Time `gorm:"default null"`
	UpdateTime  *time.Time `gorm:"autoUpdateTime;default null"`
}

func (t *Teacher) TableName() string {
	return "teachers"
}
