package lock

import (
	"context"
	"errors"
	"fmt"
	"github.com/smartystreets/goconvey/convey"
	"lottery_single/internal/pkg/middlewares/cache"
	"sync"
	"testing"
	"time"
)

func TestLock(t *testing.T) {
	// 先初始化
	cache.Init(cache.WithAddr("127.0.0.1:6379"),
		cache.WithDB(0),
		cache.WithPoolSize(10),
		cache.WithPassWord(""),
	)

	cache.GetRedisCli()

	convey.Convey("TestLock", t, func() {
		fmt.Println(5 * time.Second / 4)
		ctx := context.Background()
		// 过期时间为1秒，并开启续期
		lock1 := NewRedisLock("test_key", WithExpireSeconds(5), WithWatchDogMode())
		lock2 := NewRedisLock("test_key", WithExpireSeconds(1), WithBlock(), WithBlockWaitingSeconds(10))
		fmt.Println(lock1)
		fmt.Println(lock2)
		if err := lock1.Lock(ctx); err != nil {
			t.Error(err)
			return
		}

		go func() {
			time.Sleep(6 * time.Second)
			if err := lock1.Unlock(ctx); err != nil {
				t.Error()
			}
		}()

		if err := lock2.Lock(ctx); err != nil {
			t.Error(err)
		}

		time.Sleep(5 * time.Second)
	})
}

func TestBlockingLock(t *testing.T) {
	convey.Convey("TestBlockingLock", t, func() {

		// 过期时间为1秒
		lock1 := NewRedisLock("test_key", WithExpireSeconds(1))
		// 阻塞等待时间为2秒
		lock2 := NewRedisLock("test_key", WithBlock(), WithBlockWaitingSeconds(2))

		ctx := context.Background()
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := lock1.Lock(ctx); err != nil {
				t.Error(err)
				return
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := lock2.Lock(ctx); err != nil {
				t.Error(err)
				return
			}
		}()

		wg.Wait()

		t.Log("success")
	})
}

func Test_nonblockingLock(t *testing.T) {

	lock1 := NewRedisLock("test_key", WithExpireSeconds(1))
	lock2 := NewRedisLock("test_key")

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := lock1.Lock(ctx); err != nil {
			t.Error(err)
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := lock2.Lock(ctx); err == nil || !errors.Is(err, ErrLockAcquiredByOthers) {
			t.Errorf("got err: %v, expect: %v", err, ErrLockAcquiredByOthers)
			return
		}
	}()

	wg.Wait()
	t.Log("success")
}
