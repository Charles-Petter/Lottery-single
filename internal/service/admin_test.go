package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"io/ioutil"
	"lottery_single/configs"
	"lottery_single/internal/model"
	"lottery_single/internal/pkg/constant"
	"lottery_single/internal/pkg/middlewares/cache"
	"lottery_single/internal/pkg/middlewares/gormcli"
	"lottery_single/internal/pkg/middlewares/log"
	"lottery_single/internal/pkg/utils"
	"net/http"
	"testing"
	"time"
)

func InitTest() {
	conf := configs.InitConfig()
	logConf := conf.LogConfig
	dbConf := conf.DbConfig
	cacheConf := conf.RedisConfig

	// 初始化日志
	log.Init(
		log.WithFileName(logConf.FileName),
		log.WithLogLevel(logConf.Level),
		log.WithLogPath(logConf.LogPath),
		log.WithMaxSize(logConf.MaxSize),
		log.WithMaxBackups(logConf.MaxBackups))

	// 初始化DB
	gormcli.Init(
		gormcli.WithAddr(dbConf.Addr),
		gormcli.WithUser(dbConf.User),
		gormcli.WithPassword(dbConf.Password),
		gormcli.WithDataBase(dbConf.DataBase),
		gormcli.WithMaxIdleConn(dbConf.MaxIdleConn),
		gormcli.WithMaxOpenConn(dbConf.MaxOpenConn),
		gormcli.WithMaxIdleTime(dbConf.MaxIdleTime))

	cache.Init(
		cache.WithAddr(cacheConf.Addr),
		cache.WithPassWord(cacheConf.PassWord),
		cache.WithDB(cacheConf.DB),
		cache.WithPoolSize(cacheConf.PoolSize))

	// 初始化各个service
	Init()
}

func TestTimeFormat(t *testing.T) {
	now := time.Now()
	year := now.Year()     //年
	month := now.Month()   //月
	day := now.Day()       //日
	hour := now.Hour()     //小时
	minute := now.Minute() //分钟
	second := now.Second() //秒
	timeStr := fmt.Sprintf("%d%d%d%d%d%d", year, int(month), day, hour, minute, second)
	t.Logf("Time format: %s", timeStr)
}

func TestFormatTime(t *testing.T) {
	now := time.Now()
	i, h, m := 1, 1, 1
	dayTimeStamp := int(now.Unix()) + i*86400
	hourTimeStamp := dayTimeStamp + h*3600
	minuteTimeStamp := hourTimeStamp + m*60
	timeStr := utils.FormatFromUnixTime(int64(minuteTimeStamp))
	t.Logf("Time format: %s", timeStr)

}

func TestAddPrize(t *testing.T) {
	client := &http.Client{}
	data := make(map[string]interface{})
	data["name"] = "zhaofan"
	data["age"] = "23"

	addPrizereq := ViewPrize{
		Title:     "iphone",
		Img:       "https://p0.ssl.qhmsg.com/t016ff98b934914aca6.png",
		PrizeNum:  10,
		PrizeCode: "0-9999",
		EndTime:   time.Now().Add(time.Hour * 24 * 7),
		BeginTime: time.Now(),
		PrizeType: constant.PrizeTypeEntityLarge,
	}

	bytesData, err := json.Marshal(&addPrizereq)
	if err != nil {
		t.Errorf("Error marshalling:%v", err)
	}
	t.Logf("req json = %s\n", string(bytesData))
	req, _ := http.NewRequest("POST", "http://httpbin.org/post", bytes.NewReader(bytesData))
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func TestCreateUserAutoCreateTime(t *testing.T) {
	InitTest()
	prize1 := ViewPrize{
		Title:     "iphone",
		Img:       "https://doc-fd.zol-img.com.cn/t_s500x2000/g7/M00/05/02/ChMkK2SKcW2IGR_FAAB_sZAhtQoAARTfQPd3OsAAH_J205.jpg",
		PrizeNum:  10,
		PrizeCode: "1-10",
		EndTime:   time.Now().Add(time.Hour * 24 * 7),
		BeginTime: time.Now(),
		PrizeType: constant.PrizeTypeEntityLarge,
	}
	if err := GetAdminService().AddPrize(context.Background(), &prize1); err != nil {
		t.Logf("TestCreateUserAutoCreateTime err:%v", err)
	}

	prize2 := ViewPrize{
		Title:     "homepod",
		Img:       "https://imgservice.suning.cn/uimg1/b2c/image/t_QerWgoH9ergm0_NY4WhA.png_800w_800h_4e",
		PrizeNum:  50,
		PrizeCode: "100-150",
		EndTime:   time.Now().Add(time.Hour * 24 * 7),
		BeginTime: time.Now(),
		PrizeType: constant.PrizeTypeEntityMiddle,
	}
	if err := GetAdminService().AddPrize(context.Background(), &prize2); err != nil {
		t.Logf("TestCreateUserAutoCreateTime err:%v", err)
	}

	prize3 := ViewPrize{
		Title:     "充电器",
		Img:       "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcS5Y7iNXpcthgJ6yvE3Os1bTwLARVnvwYXeKA&usqp=CAU",
		PrizeNum:  50,
		PrizeCode: "1500-2000",
		EndTime:   time.Now().Add(time.Hour * 24 * 7),
		BeginTime: time.Now(),
		PrizeType: constant.PrizeTypeEntitySmall,
	}
	if err := GetAdminService().AddPrize(context.Background(), &prize3); err != nil {
		t.Logf("TestCreateUserAutoCreateTime err:%v", err)
	}

	prize4 := ViewPrize{
		Title:     "优惠券",
		Img:       "https://static.699pic.com/images/diversion/d66d647c52cd66beb800ba09748ea080.jpg",
		PrizeNum:  50,
		PrizeCode: "3000-6000",
		EndTime:   time.Now().Add(time.Hour * 24 * 7),
		BeginTime: time.Now(),
		PrizeType: constant.PrizeTypeCouponDiff,
	}
	if err := GetAdminService().AddPrize(context.Background(), &prize4); err != nil {
		t.Logf("TestCreateUserAutoCreateTime err:%v", err)
	}
}

func TestImportCoupon(t *testing.T) {
	InitTest()
	couponInfo := ViewCouponInfo{
		PrizeId: 20,
		Code: "coupon_code0000001\n" +
			"coupon_code0000002\n" +
			"coupon_code0000003\n" +
			"coupon_code0000004\n" +
			"coupon_code0000005",
		SysCreated: time.Time{},
		SysUpdated: time.Time{},
		SysStatus:  1, // 正常
	}
	successNum, failNum, err := GetAdminService().ImportCoupon(context.Background(), couponInfo.PrizeId, couponInfo.Code)
	if err != nil {
		t.Errorf("TestImportCoupon|ImportCoupon err: %v", err)
	}
	t.Logf("successNum=%d, failNum=%d\n", successNum, failNum)
}

func TestImportCouponWithCache(t *testing.T) {
	InitTest()
	couponInfo := ViewCouponInfo{
		PrizeId: 20,
		Code: "coupon_code0000001\n" +
			"coupon_code0000002\n" +
			"coupon_code0000003\n" +
			"coupon_code0000004\n" +
			"coupon_code0000005",
		SysStatus: 1, // 正常
	}
	successNum, failNum, err := GetAdminService().ImportCouponWithCache(context.Background(), couponInfo.PrizeId, couponInfo.Code)
	if err != nil {
		t.Errorf("TestImportCoupon|ImportCoupon err: %v", err)
	}
	t.Logf("successNum=%d, failNum=%d\n", successNum, failNum)
}

func TestAddUser(t *testing.T) {
	InitTest()
	pwd := []byte("123456")
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	if err != nil {
		t.Errorf("bcrypt.GenerateFromPassword error: %v", err)
	}
	t.Logf("hash=%s\n", hash)

	user1 := model.User{
		UserName: "zhangsan",
		Password: string(hash),
		Age:      30,
		Gender:   "male", // 设置性别为男性
	}
	if err := GetAdminService().AddUser(context.Background(), &user1); err != nil {
		t.Logf("TestAddUser|AddUser err: %v", err)
	}
	user2 := model.User{
		UserName: "lisi",
		Password: string(hash),
		Age:      25,
		Gender:   "female", // 设置性别为女性
	}
	if err := GetAdminService().AddUser(context.Background(), &user2); err != nil {
		t.Logf("TestAddUser|AddUser err: %v", err)
	}
}

func TestLogin(t *testing.T) {
	InitTest()
	userName := "zhangsan"
	pwd := "123456"
	loginRsp, err := GetUserService().Login(context.Background(), userName, pwd)
	if err != nil {
		t.Errorf("TestLogin|Login err: %v", err)
	}
	t.Logf("TestLogin|Login rsp: %v", loginRsp)
}

func TestUpdatePrize(t *testing.T) {
	InitTest()
	code := "2001-9999"
	prize4 := model.Prize{
		Id:        20,
		Title:     "优惠券",
		Img:       "https://static.699pic.com/images/diversion/d66d647c52cd66beb800ba09748ea080.jpg",
		PrizeNum:  50,
		PrizeCode: code,
		EndTime:   time.Now().Add(time.Hour * 24 * 7),
		BeginTime: time.Now(),
		PrizeType: constant.PrizeTypeCouponDiff,
	}
	if err := GetAdminService().UpdateDbPrize(context.Background(), gormcli.GetDB(), &prize4, "prize_code"); err != nil {
		t.Logf("TestUpdatePrize err:%v", err)
	}
}

func TestExpr(t *testing.T) {
	InitTest()
	res := gormcli.GetDB().Model(&model.Prize{}).Where("id = ? and left_num >= ?", 20, 1).UpdateColumn("left_num", gorm.Expr("left_num+ + ?", 1))
	if res.Error != nil {
		t.Errorf("expr err:%v", res.Error)
	}
	if res.RowsAffected <= 0 {
		t.Log(1111111)
	}
}

func TestGetAllPrizeByCache(t *testing.T) {
	InitTest()
	valutStr, ok, err := cache.GetRedisCli().Get(context.Background(), constant.AllPrizeCacheKey)
	t.Log(valutStr)
	t.Log(ok)
	t.Log(err)
}

func TestTimeHour(t *testing.T) {
	t.Log(time.Now().Hour())
}
func TestGetPrizeList(t *testing.T) {
	InitTest()

	// 创建一个context
	ctx := context.Background()

	// 获取AdminService
	adminService := GetAdminService()

	// 调用GetPrizeList方法
	prizes, err := adminService.GetPrizeList(ctx)

	// 使用assert进行错误检查和结果验证
	assert.NoError(t, err)
	assert.NotNil(t, prizes)
}
