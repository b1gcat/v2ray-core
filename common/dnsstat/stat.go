package dnsstat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Dns struct {
	TopN           int    `json:"topN"`
	Mac            string `json:"router_mac"`
	Retention      int    `json:"retention"`       //保留最近多久的记录hour
	Path           string `json:"db_path"`         //数据库路径
	ApiURL         string `json:"api_url"`         //api令牌
	ReportInterval int    `json:"report_interval"` //上报周期s

	Ctx    context.Context
	Logger *logrus.Logger

	valid    bool
	pipe     chan string
	dbLocker sync.Mutex
	db       *gorm.DB
	last     time.Time
}

type Data struct {
	Timestamp time.Time `json:"timestamp"`
	Site      string    `json:"site"`
	Count     int       `json:"count"`
}

var DnsStat Dns

func Stat() *Dns {
	return &DnsStat
}

func (d *Dns) Run() error {
	db, err := gorm.Open(sqlite.Open(d.Path))
	if err != nil {
		return err
	}

	if err := db.AutoMigrate(&Data{}); err != nil {
		return err
	}

	d.db = db
	d.pipe = make(chan string, 2048)
	go d.runCleanUp()
	go d.runDb()

	d.valid = true
	d.runReport()
	return nil
}

func (d *Dns) Insert(domain string) error {
	if !d.valid {
		return nil
	}

	domain = strings.TrimSuffix(domain, ".")
	dm := strings.Split(domain, ".")
	if len(dm) > 2 {
		domain = dm[len(dm)-2] + "." + dm[len(dm)-1]
	}

	select {
	case d.pipe <- domain:
	default:
		d.Logger.Errorf("full queue: %v(failed to statistics)", domain)
	}
	return nil

}

func (d *Dns) runDb() {
	d.dbLocker.Lock()

	for {
		d.dbLocker.Unlock()

		domain := <-d.pipe

		d.dbLocker.Lock()

		item := Data{}
		tx := d.db.Where("site = ?", domain).Find(&item)
		if tx.Error != nil {
			d.Logger.Errorf("runDb:%v", tx.Error.Error())
			continue
		}

		item.Count++
		item.Timestamp = time.Now()
		item.Site = domain

		if tx.RowsAffected == 0 {
			if err := d.db.Create(item).Error; err != nil {
				d.Logger.Errorf("runDb.create.item:%v", tx.Error.Error())
			}
		} else {
			if err := d.db.Model(&item).
				Where("site = ?", domain).
				Updates(&item).Error; err != nil {
				d.Logger.Errorf("runDb.update.item:%v", tx.Error.Error())
			}
		}

	}
}

type Report struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	RouterMac string    `json:"router_mac"`

	Record []Data `json:"record"`
}

type apiMessage struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (d *Dns) runReport() {
	t := time.NewTicker(time.Second * time.Duration(d.ReportInterval))
	defer t.Stop()

	for {

		//从定时器中获取数据
		select {
		case <-t.C:
		case <-d.Ctx.Done():
			return
		}
		var items []Data

		d.dbLocker.Lock()
		tx := d.db.Order("count desc").Limit(d.TopN).Find(&items)
		if tx.Error != nil {
			d.dbLocker.Unlock()
			d.Logger.Error("runReport:", tx.Error.Error())
			continue
		}
		d.dbLocker.Unlock()

		if tx.RowsAffected == 0 {
			continue
		}

		if d.last.IsZero() {
			d.last = time.Now()
		}

		data := &Report{
			RouterMac: d.Mac,
			StartTime: d.last,
			EndTime:   time.Now(),
			Record:    items,
		}

		aJson, err := json.Marshal(data)
		if err != nil {
			d.Logger.Error("runReport.Marshal:", err.Error())
			continue
		}

		if err = d.postData(aJson); err != nil {
			d.Logger.Error("runReport.postData:", err.Error())
			continue
		}

		d.dbLocker.Lock()
		d.last = time.Now()
		if err := d.db.Model(&Data{}).Where("1=1").Delete(&Data{}).Error; err != nil {
			d.Logger.Error("[x] runReport.delete.clean:", err.Error())
		}
		d.dbLocker.Unlock()
	}
}

func (d *Dns) postData(data []byte) error {
	hc := &http.Client{Timeout: 10 * time.Second}
	resp, err := hc.Post(d.ApiURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("服务器内部错误:%s", resp.Status)
	}

	rBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	r := apiMessage{}
	if err = json.Unmarshal(rBytes, &r); err != nil {
		return fmt.Errorf("通过API获取的格式:%v", err.Error())
	}

	if r.ErrCode != 0 {
		return fmt.Errorf("API返回错误码:%v:%s", r.ErrCode, r.ErrMsg)
	}

	return nil
}

func (d *Dns) runCleanUp() {
	t := time.NewTicker(time.Hour * time.Duration(d.Retention))
	defer t.Stop()

	for {

		//从定时器中获取数据
		select {
		case <-t.C:
		case <-d.Ctx.Done():
			return
		}
		clearTime, err := time.ParseDuration(fmt.Sprintf(
			"-%dh", d.Retention))
		if err != nil {
			d.Logger.Error("[x] dnsstat.runCleanUp.ParseDuration:", err.Error())
			return
		}
		timeWaterMaker := time.Now().Add(clearTime).Format("2006-01-02 15:04:05")
		d.Logger.Infof("清理dns统计 timestamp < %s", timeWaterMaker)

		//清理记录
		d.dbLocker.Lock()
		if err := d.db.Unscoped().Where("timestamp < ?", timeWaterMaker).Delete(&Data{}).Error; err != nil {
			d.Logger.Error("[x] dnsstat.runCleanUp.clean:", err.Error())
		}
		d.dbLocker.Unlock()
	}
}
