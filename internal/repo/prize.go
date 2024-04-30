package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"lottery_single/internal/model"
	"lottery_single/internal/pkg/constant"
	"lottery_single/internal/pkg/middlewares/cache"
	"lottery_single/internal/pkg/middlewares/log"
	"lottery_single/internal/pkg/utils"
	"strconv"
	"time"
)

type PrizeReop struct {
}

func NewPrizeRepo() *PrizeReop {
	return &PrizeReop{}
}

func (r *PrizeReop) Get(db *gorm.DB, id uint) (*model.Prize, error) {
	prize := &model.Prize{
		Id: id,
	}
	err := db.Model(&model.Prize{}).First(prize).Error
	if err != nil {
		if err.Error() == gorm.ErrRecordNotFound.Error() {
			return nil, nil
		}
		return nil, fmt.Errorf("PrizeRepo|Get:%v", err)
	}
	return prize, nil
}

func (r *PrizeReop) GetWithCache(db *gorm.DB, id uint) (*model.Prize, error) {
	prizeList, err := r.GetAllWithCache(db)
	if err != nil {
		return nil, fmt.Errorf("PrizeRepo|GetWithCache:%v", err)
	}
	for _, prize := range prizeList {
		if prize.Id == id {
			return prize, nil
		}
	}
	return nil, nil
}

func (r *PrizeReop) GetAll(db *gorm.DB) ([]*model.Prize, error) {
	var prizes []*model.Prize
	err := db.Model(&model.Prize{}).Find(&prizes).Error
	if err != nil {
		return nil, fmt.Errorf("PrizeRepo|GetAll:%v", err)
	}
	return prizes, nil
}

func (r *PrizeReop) GetAllWithCache(db *gorm.DB) ([]*model.Prize, error) {
	prizeList, err := r.GetAllByCache()
	if err != nil {
		return nil, fmt.Errorf("PrizeRepo|GetAllWithCache:%v", err)
	}
	if prizeList == nil {
		// 缓存没查到，从db获取
		prizeList, err = r.GetAll(db)
		if err != nil {
			return nil, fmt.Errorf("PrizeRepo|GetAllWithCache:%v", err)
		}
		// 将数据更新到缓存中
		if err = r.SetAllByCache(prizeList); err != nil {
			return nil, fmt.Errorf("PrizeRepo|GetAllWithCache:%v", err)
		}
	}
	return prizeList, nil
}

func (r *PrizeReop) CountAll(db *gorm.DB) (int64, error) {
	var num int64
	err := db.Model(&model.Prize{}).Count(&num).Error
	if err != nil {
		return 0, fmt.Errorf("PrizeRepo|CountAll:%v", err)
	}
	return num, nil
}

func (r *PrizeReop) CountAllWithCache(db *gorm.DB) (int64, error) {
	prizeList, err := r.GetAllWithCache(db)
	if err != nil {
		return 0, fmt.Errorf("PrizeRepo|CountAllWithCache:%v", err)
	}
	return int64(len(prizeList)), nil
}

func (r *PrizeReop) Create(db *gorm.DB, prize *model.Prize) error {
	err := db.Model(&model.Prize{}).Create(prize).Error
	if err != nil {
		return fmt.Errorf("PrizeRepo|Create:%v", err)
	}
	return nil
}

func (r *PrizeReop) CreateWithCache(db *gorm.DB, prize *model.Prize) error {
	if err := r.UpdateByCache(prize); err != nil {
		return fmt.Errorf("PrizeRepo|CreateWithCache:%v", err)
	}
	return r.Create(db, prize)
}

func (r *PrizeReop) Delete(db *gorm.DB, id uint) error {
	prize := &model.Prize{Id: id}
	if err := db.Model(&model.Prize{}).Delete(prize).Error; err != nil {
		return fmt.Errorf("PrizeRepo|Delete:%v")
	}
	return nil
}

func (r *PrizeReop) DeleteWithCache(db *gorm.DB, id uint) error {
	prize := &model.Prize{
		Id: id,
	}
	if err := r.UpdateByCache(prize); err != nil {
		return fmt.Errorf("PrizeRepo|DeleteWithCache:%v", err)
	}
	return r.Delete(db, id)
}

func (r *PrizeReop) Update(db *gorm.DB, prize *model.Prize, cols ...string) error {
	var err error
	fmt.Printf("PrizeRepo|Update=%+v\n", prize)
	if len(cols) == 0 {
		err = db.Model(prize).Updates(prize).Error
	} else {
		err = db.Model(prize).Select(cols).Updates(prize).Error
	}
	if err != nil {
		return fmt.Errorf("PrizeRepo|Update:%v", err)
	}
	return nil
}

func (r *PrizeReop) UpdateWithCache(db *gorm.DB, prize *model.Prize, cols ...string) error {
	if err := r.UpdateByCache(prize); err != nil {
		return fmt.Errorf("PrizeRepo|UpdateWithCache:%v", err)
	}
	return r.Update(db, prize, cols...)
}

// GetFromCache 根据id从缓存获取奖品
func (r *PrizeReop) GetFromCache(id uint) (*model.Prize, error) {
	redisCli := cache.GetRedisCli()
	idStr := strconv.FormatUint(uint64(id), 10)
	ret, exist, err := redisCli.Get(context.Background(), idStr)
	if err != nil {
		log.Errorf("PrizeRepo|GetFromCache:" + err.Error())
		return nil, err
	}

	if !exist {
		return nil, nil
	}

	prize := model.Prize{}
	json.Unmarshal([]byte(ret), &model.Prize{})

	return &prize, nil
}

func (r *PrizeReop) GetAllUsefulPrizeList(db *gorm.DB) ([]*model.Prize, error) {
	now := time.Now()
	list := make([]*model.Prize, 0)
	err := db.Model(&model.Prize{}).Where("begin_time<=?", now).Where("end_time >= ?", now).
		Where("prize_num>?", 0).Where("sys_status=?", 1).Order("sys_updated desc").
		Order("display_order asc").Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("PrizeRepo|GetAllUsefulPrizeList:%v", err)
	}
	return list, nil
}

// GetAllUsefulPrizeListWithCache 筛选出符合条件的奖品列表
func (r *PrizeReop) GetAllUsefulPrizeListWithCache(db *gorm.DB) ([]*model.Prize, error) {
	// 优先从缓存取，缓存没取到，从db取
	prizeList, err := r.GetAllWithCache(db)
	if err != nil {
		return nil, fmt.Errorf("PrizeRepo|GetAllUsefulPrizeListWithCache:%v", err)
	}
	now := time.Now()
	dataList := make([]*model.Prize, 0)
	for _, prize := range prizeList {
		if prize.Id > 0 && prize.SysStatus == 1 && prize.PrizeNum > 0 &&
			prize.BeginTime.Before(now) && prize.EndTime.After(now) {
			dataList = append(dataList, prize)
		}
	}
	return dataList, nil
}

func (r *PrizeReop) DecrLeftNum(db *gorm.DB, id int, num int) (bool, error) {
	log.Infof("id: %d, num: %d\n", id, num)
	res := db.Model(&model.Prize{}).Where("id = ? and left_num >= ?", id, num).UpdateColumn("left_num", gorm.Expr("left_num - ?", num))
	if res.Error != nil {
		return false, fmt.Errorf("PrizeRepo|DecrLeftNum:%v", res.Error)
	}
	if res.RowsAffected <= 0 {
		return false, nil
	}
	return true, nil
}

// DecrLeftNumByPool 奖品缓冲池 对应奖品数量递减
func (r *PrizeReop) DecrLeftNumByPool(prizeID int) (int64, error) {
	key := constant.PrizePoolCacheKey
	field := strconv.Itoa(prizeID)
	cnt, err := cache.GetRedisCli().HIncrBy(context.Background(), key, field, -1)
	if err != nil {
		return -1, fmt.Errorf("PrizeRepo|DecrLeftNumByPool:%v", err)
	}
	return cnt, nil
}

func (r *PrizeReop) IncrLeftNum(db *gorm.DB, id int, column string, num int) error {
	if err := db.Model(&model.Prize{}).Where("id = ?", id).
		Update(column, gorm.Expr(column+" + ？", num)).Error; err != nil {
		return fmt.Errorf("PrizeRepo|IncrLeftNum err: %v", err)
	}
	return nil
}

// SetAllByCache 全量数据保存到redis中
func (r *PrizeReop) SetAllByCache(prizeList []*model.Prize) error {
	value := ""
	if len(prizeList) > 0 {
		prizeMapList := make([]map[string]interface{}, len(prizeList))
		for i := 0; i < len(prizeList); i++ {
			prize := prizeList[i]
			prizeMap := make(map[string]interface{})
			prizeMap["Id"] = prize.Id
			prizeMap["Title"] = prize.Title
			prizeMap["PrizeNum"] = prize.PrizeNum
			prizeMap["LeftNum"] = prize.LeftNum
			prizeMap["PrizeCode"] = prize.PrizeCode
			prizeMap["PrizeTime"] = prize.PrizeTime
			prizeMap["Img"] = prize.Img
			prizeMap["DisplayOrder"] = prize.DisplayOrder
			prizeMap["PrizeType"] = prize.PrizeType
			prizeMap["PrizeProfile"] = prize.PrizeProfile
			prizeMap["BeginTime"] = utils.FormatFromUnixTime(prize.BeginTime.Unix())
			prizeMap["EndTime"] = utils.FormatFromUnixTime(prize.EndTime.Unix())
			//prizeMap["PrizePlan"] = prize.PrizePlan
			prizeMap["PrizeBegin"] = utils.FormatFromUnixTime(prize.PrizeBegin.Unix())
			prizeMap["PrizeEnd"] = utils.FormatFromUnixTime(prize.PrizeEnd.Unix())
			prizeMap["SysStatus"] = prize.SysStatus
			prizeMap["SysCreated"] = utils.FormatFromUnixTime(prize.SysCreated.Unix())
			prizeMap["SysUpdated"] = utils.FormatFromUnixTime(prize.SysUpdated.Unix())
			prizeMap["SysIp"] = prize.SysIp
			prizeMapList[i] = prizeMap
		}
		bytes, err := json.Marshal(prizeMapList)
		if err != nil {
			log.Errorf("SetAllByCache|marshal err:%v", err)
			return fmt.Errorf("SetAllByCache|marshal err:%v", err)
		}
		value = string(bytes)
	}
	if err := cache.GetRedisCli().Set(context.Background(), constant.AllPrizeCacheKey, value, time.Second*time.Duration(constant.AllPrizeCacheTime)); err != nil {
		log.Errorf("SetAllByCache|set cache err:%v", err)
		return fmt.Errorf("SetAllByCache|set cache err:%v", err)
	}
	return nil
}

// GetAllByCache 从缓存中获取所有的奖品信息
func (r *PrizeReop) GetAllByCache() ([]*model.Prize, error) {
	valutStr, ok, err := cache.GetRedisCli().Get(context.Background(), constant.AllPrizeCacheKey)
	if err != nil {
		return nil, fmt.Errorf("PrizeRepo|GetAllByCache:%v", err)
	}
	// 缓存中没数据
	if !ok {
		return nil, nil
	}
	str := utils.GetString(valutStr, "")
	if str == "" {
		return nil, nil
	}
	// 将json数据反序列化
	prizeMapList := []map[string]interface{}{}
	err = json.Unmarshal([]byte(str), &prizeMapList)
	if err != nil {
		log.Errorf("PrizeRepo|GetAllByCache:%v", err)
		return nil, fmt.Errorf("PrizeRepo|GetAllByCache:%v", err)
	}
	prizeList := make([]*model.Prize, len(prizeMapList))
	for i := 0; i < len(prizeMapList); i++ {
		prizeMap := prizeMapList[i]
		id := utils.GetInt64FromMap(prizeMap, "Id", 0)
		if id <= 0 {
			prizeList[i] = &model.Prize{}
			continue
		}
		prizeBegin, err := utils.ParseTime(utils.GetStringFromMap(prizeMap, "PrizeBegin", ""))
		if err != nil {
			log.Errorf("PrizeRepo|GetAllByCache ParseTime PrizeBegin err:%v", err)
			return nil, fmt.Errorf("PrizeRepo|GetAllByCache:%v", err)
		}
		prizeEnd, err := utils.ParseTime(utils.GetStringFromMap(prizeMap, "PrizeEnd", ""))
		if err != nil {
			log.Errorf("PrizeRepo|GetAllByCache ParseTime PrizeEnd err:%v", err)
			return nil, fmt.Errorf("PrizeRepo|GetAllByCache:%v", err)
		}
		beginTime, err := utils.ParseTime(utils.GetStringFromMap(prizeMap, "BeginTime", ""))
		if err != nil {
			log.Errorf("PrizeRepo|GetAllByCache ParseTime BeginTime err:%v", err)
			return nil, fmt.Errorf("PrizeRepo|GetAllByCache:%v", err)
		}
		endTime, err := utils.ParseTime(utils.GetStringFromMap(prizeMap, "EndTime", ""))
		if err != nil {
			log.Errorf("PrizeRepo|GetAllByCache ParseTime EndTime err:%v", err)
			return nil, fmt.Errorf("PrizeRepo|GetAllByCache:%v", err)
		}
		sysCreated, err := utils.ParseTime(utils.GetStringFromMap(prizeMap, "SysCreated", ""))
		if err != nil {
			log.Errorf("PrizeRepo|GetAllByCache ParseTime SysCreated err:%v", err)
			return nil, fmt.Errorf("PrizeRepo|GetAllByCache:%v", err)
		}
		sysUpdated, err := utils.ParseTime(utils.GetStringFromMap(prizeMap, "SysUpdated", ""))
		if err != nil {
			log.Errorf("PrizeRepo|GetAllByCache ParseTime SysUpdated err:%v", err)
			return nil, fmt.Errorf("PrizeRepo|GetAllByCache:%v", err)
		}
		prize := &model.Prize{
			Id:           uint(id),
			Title:        utils.GetStringFromMap(prizeMap, "Title", ""),
			PrizeNum:     int(utils.GetInt64FromMap(prizeMap, "PrizeNum", 0)),
			LeftNum:      int(utils.GetInt64FromMap(prizeMap, "LeftNum", 0)),
			PrizeCode:    utils.GetStringFromMap(prizeMap, "PrizeCode", ""),
			PrizeTime:    uint(utils.GetInt64FromMap(prizeMap, "PrizeTime", 0)),
			Img:          utils.GetStringFromMap(prizeMap, "Img", ""),
			DisplayOrder: uint(utils.GetInt64FromMap(prizeMap, "DisplayOrder", 0)),
			PrizeType:    uint(utils.GetInt64FromMap(prizeMap, "PrizeType", 0)),
			PrizeProfile: utils.GetStringFromMap(prizeMap, "PrizeProfile", ""),
			BeginTime:    beginTime,
			EndTime:      endTime,
			//PrizeData:    comm.GetStringFromMap(data, "PrizeData", ""),
			PrizeBegin: prizeBegin,
			PrizeEnd:   prizeEnd,
			SysStatus:  uint(utils.GetInt64FromMap(prizeMap, "SysStatus", 0)),
			SysCreated: &sysCreated,
			SysUpdated: &sysUpdated,
			SysIp:      utils.GetStringFromMap(prizeMap, "SysIp", ""),
		}
		prizeList[i] = prize
	}
	return prizeList, nil
}

// UpdateByCache 数据更新，需要更新缓存，直接清空缓存数据
func (r *PrizeReop) UpdateByCache(prize *model.Prize) error {
	if prize == nil || prize.Id <= 0 {
		return nil
	}
	if err := cache.GetRedisCli().Delete(context.Background(), constant.AllPrizeCacheKey); err != nil {
		return fmt.Errorf("PrizeRepo|UpdateByCache err:%v", err)
	}
	return nil
}

// GetPrizePoolNum 获取奖品缓冲池中获取数据
func (r *PrizeReop) GetPrizePoolNum(prizeID uint) (int, error) {
	key := constant.PrizePoolCacheKey
	field := strconv.Itoa(int(prizeID))
	res, err := cache.GetRedisCli().HGet(context.Background(), key, field)
	if err != nil {
		return 0, fmt.Errorf("PrizeRepo|GetPrizePoolNum:%v", err)
	}
	num, err := strconv.Atoi(res)
	if err != nil {
		return 0, fmt.Errorf("PrizeRepo|GetPrizePoolNum:%v", err)
	}
	return num, nil
}
