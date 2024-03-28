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

	valid bool
	lock  sync.RWMutex
	db    *gorm.DB
	last  time.Time
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

	go d.runCleanUp()

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

	//search
	d.lock.Lock()
	defer d.lock.Unlock()

	item := Data{}
	tx := d.db.Where("site = ?", domain).Find(&item)
	if tx.Error != nil {
		return tx.Error
	}
	item.Count++
	item.Timestamp = time.Now()
	item.Site = domain

	if tx.RowsAffected == 0 {
		return d.db.Create(item).Error
		//无记录
	} else {
		return d.db.Model(&item).Where("site = ?", domain).Updates(&item).Error
	}
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

		d.lock.Lock()
		err := d.db.Order("count desc").Limit(d.TopN).Find(&items).Error
		if err != nil {
			d.lock.Unlock()
			d.Logger.Error("runReport:", err.Error())
			continue
		}
		d.lock.Unlock()

		type Report struct {
			StartTime time.Time `json:"start_time"`
			EndTime   time.Time `json:"end_time"`
			RouterMac string    `json:"router_mac"`

			Record []Data `json:"record"`
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
		d.last = time.Now()
		d.lock.Lock()
		err = d.db.Model(&Data{}).Where("1=1").Update("count", 0).Error
		d.Logger.Infof("runReport.reset.counter:%v", err)
		d.lock.Unlock()
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

	type apiMessage struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
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
	t := time.NewTicker(time.Hour)
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
		d.lock.Lock()
		defer d.lock.Unlock()
		if err := d.db.Unscoped().Where("timestamp < ?", timeWaterMaker).Delete(&Data{}).Error; err != nil {
			d.Logger.Error("dnsstat.runCleanUp.Delete:", err.Error())
		}
	}
}
