package task

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"sync"
	"time"

	"github.com/kouleen/lets-encrypt/internal/modle"
	"github.com/kouleen/lets-encrypt/internal/service"
	"github.com/kouleen/lets-encrypt/pkg/util"
	"github.com/kouleen/lets-encrypt/static"
	"github.com/robfig/cron/v3"
)

var mu sync.Mutex

var tpl *template.Template

func init() {
	noticeTpl, err := template.ParseFS(static.FS, "notice.html")
	if err != nil {
		fmt.Printf("读取模板失败: %v\n", err)
		return
	}
	tpl = noticeTpl
	// WithSeconds()开启秒解析；必须设置本地时区，否则默认UTC时区！
	c := cron.New(cron.WithSeconds(), cron.WithLocation(time.Local))
	// 每天执行一次
	//_, err = c.AddFunc("0 */1 * * * *", dailyJob)		// 每分钟执行一次 测试
	if _, err = c.AddFunc("0 0 0 * * *", dailyJob); err != nil {
		log.Fatal("job init error: %w", err)
	}
	c.Start()
	log.Println("cron启动，等待执行...")
}

func dailyJob() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("job panic: %v", err)
		}
		mu.Unlock()
	}()
	if !mu.TryLock() {
		log.Println("上一轮未结束，跳过本次任务")
		return
	}
	log.Printf("开始执行任务 %s", time.Now().Format("2006-01-02 15:04:05"))
	now := time.Now()
	autoRefreshReload(now)
	noticeEmail(now)
}

// autoRefreshReload 自动刷新
func autoRefreshReload(now time.Time) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("autoRefreshReload panic: %v", err)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	// 先查询出数据
	v := uint8(1)
	req := &modle.AcmeEncryptQuery{
		Status: &v,
		Auto:   &v,
	}
	encryptList, err := service.ListAcmeEncrypt(ctx, req)
	if err != nil {
		log.Printf("dailyJob job panic: %v", err)
		return
	}
	for _, encrypt := range encryptList {
		// 过滤掉证书路径不存在的情况
		if encrypt.Encrypt == "" {
			continue
		}
		expireTime := encrypt.ExpireTime
		remainDur := expireTime.Sub(now)
		remainDay := int(remainDur.Hours() / 24)
		// 过滤掉没有过续期阈值的数据
		if remainDay > encrypt.RemainDay {
			continue
		}
		_, err = service.RefreshAcmeEncrypt(ctx, encrypt.GetDomain())
		if err != nil {
			log.Printf("RefreshAcmeEncrypt job domain: %s, panic: %v", encrypt.GetDomain(), err)
		}
	}
}

// noticeEmail 通知邮件
func noticeEmail(now time.Time) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("noticeEmail panic: %v", err)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	// 先查询出数据
	v := uint8(1)
	req := &modle.AcmeEncryptQuery{
		Status: &v,
		Notice: &v,
	}
	encryptList, err := service.ListAcmeEncrypt(ctx, req)
	if err != nil {
		log.Printf("dailyJob job panic: %v", err)
		return
	}
	noticeEmailMap := make(map[string][]*NoticeEmail)
	for _, encrypt := range encryptList {

		expireTime := encrypt.ExpireTime
		remainDur := expireTime.Sub(now)
		remainDay := int(remainDur.Hours() / 24)
		// 已经过期的不通知
		if remainDay <= 0 {
			continue
		}
		// 过滤掉没有过续期阈值的数据
		if remainDay > encrypt.RemainDay {
			continue
		}
		// 剩余天数 <= 续期阈值
		ne := &NoticeEmail{
			Domain:     encrypt.GetDomain(),
			ExpireTime: *encrypt.ExpireTime,
			RemainDay:  remainDay,
			RenewalDay: encrypt.RemainDay,
		}
		noticeEmailMap[encrypt.Username] = append(noticeEmailMap[encrypt.Username], ne)
	}

	for username, items := range noticeEmailMap {
		var buf bytes.Buffer
		if err = tpl.Execute(&buf, map[string]any{"Items": items}); err != nil {
			log.Printf("渲染邮件模板失败 noticeEmail template execute username: %s error: %v", username, err)
			return
		}
		util.SendMail(ctx, username, buf.String())
	}
}

type NoticeEmail struct {
	Domain     string    `json:"domain"`     // 域名
	ExpireTime time.Time `json:"expireTime"` // 过期时间
	RemainDay  int       `json:"remainDay"`  // 剩余阈值
	RenewalDay int       `json:"renewalDay"` // 续期阈值
}
