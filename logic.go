package main

import (
	"net/url"
	"strconv"
	"net/http"
	"log"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"bytes"
	"fmt"
	"strings"
	"github.com/xen0n/go-workwx"
	"context"
	"time"
)

type InRegistrationHouse struct {
	Name              string // 区域
	Area              string // 项目名称
	No                string // 预售证号
	Range             string // 预售范围
	Count             string // 住房套数
	ContactPhone      string // 开发商咨询电话
	RegisterStartTime string // 登记开始时间
	RegisterEndTime   string // 登记结束时间
	InReleaseTime     string // 名单内人员资格已释放时间
	OutReleaseTime    string // 名单外人员资格已释放时间
	PreAuditEndTime   string // 预审码取得截止时间
	State             string // 项目报名状态
}

func InRegistrationHouses() (houses []InRegistrationHouse, err error) {
	params := url.Values{}
	params.Set("regioncode", "00")
	pageNo := 1
	hasMore := true
	for pageNo < 10 && hasMore {
		params.Set("pageNo", strconv.Itoa(pageNo))
		func() {
			var resp *http.Response
			if resp, err = http.PostForm("https://zw.cdzj.chengdu.gov.cn/lottery/accept/projectList", params); err != nil {
				log.Println("failed to fetch in registration houses", err.Error())
				return
			}
			defer func() { _ = resp.Body.Close() }()
			var body []byte
			if body, err = ioutil.ReadAll(resp.Body); err != nil {
				log.Println("failed to read in registration houses response body", err.Error())
				return
			}
			var doc *goquery.Document
			if doc, err = goquery.NewDocumentFromReader(bytes.NewReader(body)); err != nil {
				log.Println("failed to parse in registration houses document", err.Error())
				return
			}
			index2name := make(map[int]string)
			doc.Find("thead").Find("th").Each(func(i int, selection *goquery.Selection) {
				index2name[i] = selection.Text()
			})
			doc.Find("tbody#_projectInfo").Find("tr").Each(func(_ int, p *goquery.Selection) {
				house := InRegistrationHouse{}
				p.Find("td").Each(func(i int, selection *goquery.Selection) {
					switch index2name[i] {
					case "区域":
						house.Area = selection.Text()
					case "项目名称":
						house.Name = selection.Text()
					case "预售证号":
						house.No = selection.Text()
					case "预售范围":
						house.Range = selection.Text()
					case "住房套数":
						house.Count = selection.Text()
					case "开发商咨询电话":
						house.ContactPhone = selection.Text()
					case "登记开始时间":
						house.RegisterStartTime = selection.Text()
					case "登记结束时间":
						house.RegisterEndTime = selection.Text()
					case "名单外人员资格已释放时间":
						house.OutReleaseTime = selection.Text()
					case "名单内人员资格已释放时间":
						house.InReleaseTime = selection.Text()
					case "预审码取得截止时间":
						house.PreAuditEndTime = selection.Text()
					case "项目报名状态":
						house.State = selection.Text()
					}
				})
				if house.State == "正在报名" {
					houses = append(houses, house)
				} else {
					hasMore = false
				}
			})
		}()
		pageNo++
	}
	return
}

type House struct {
	Alias string
	Price string
}

func HouseInfo(name string) (house House, err error) {
	params := url.Values{}
	params.Set("keyword", name)
	var resp *http.Response
	if resp, err = http.Get(fmt.Sprintf("https://www.cdgoufangtong.com/search?%s", params.Encode())); err != nil {
		log.Println("failed to fetch house info", name, err.Error())
		return
	}
	defer func() { _ = resp.Body.Close() }()
	var doc *goquery.Document
	if doc, err = goquery.NewDocumentFromReader(resp.Body); err != nil {
		log.Println("failed to parse in registration houses document", err.Error())
		return
	}
	doc.Find(".building-item__body").Each(func(i int, p *goquery.Selection) {
		if i == 0 {
			house.Alias = p.Find("h1").Text()
		}
	})
	doc.Find(".building-item__footer").Each(func(i int, selection *goquery.Selection) {
		if i == 0 {
			house.Price = strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(selection.Find("div").Text()), " ", ""), "\n", "")
		}
	})
	if house.Price[:len(house.Price)/2] == house.Price[len(house.Price)/2:] {
		house.Price = house.Price[:len(house.Price)/2]
	}
	return
}

func NotifyNewHouses(users string) (err error) {
	registrationHouses, err := InRegistrationHouses()
	if err != nil {
		return
	}
	for index := range registrationHouses {
		ih := &registrationHouses[index]
		//_, _ = Redis.Del(context.TODO(), ih.Name).Result()
		var notExist bool
		if notExist, err = Redis.SetNX(context.TODO(), ih.Name, 1, time.Hour*24).Result(); err != nil {
			return
		} else if !notExist {
			continue
		}
		var house House
		if house, err = HouseInfo(ih.Name); err != nil {
			return
		}
		if err = WorkWeChatApp.SendTextMessage(&workwx.Recipient{UserIDs: strings.Split(users, "|")},
			fmt.Sprintf("区域:%s\n楼盘名:%s\n别名:%s\n价格:%s\n预售证号:%s\n预售范围:%s\n住房套数:%s\n登记时间:%s\n名单内释放时间:%s\n名单外释放时间:%s",
				ih.Area, ih.Name, house.Alias, house.Price, ih.No, ih.Range, ih.Count, fmt.Sprintf("%s到%s", ih.RegisterStartTime, ih.RegisterEndTime), ih.InReleaseTime, ih.OutReleaseTime),
			false); err != nil {
			_, _ = Redis.Del(context.TODO(), ih.Name).Result()
			return
		}
	}
	return
}
