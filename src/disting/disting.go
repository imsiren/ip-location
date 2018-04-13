package disting

import (
	"github.com/fsnotify/fsnotify"
	"github.com/ipipdotnet/datx-go"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type LocalData struct {
	Country      string `json:"country"`
	Province     string `json:"province"`
	CityName     string `json:"city"`
	AreaCode     string `json:"city_code"`
	ProvinceCode string `json:"province_code"`
	CountryCode  string `json:"country_code"`
}

type Disting struct {
	DbCity        *datx.City
	DataPath      string
	CodeDataPath  string
	LatestModTime int64
	PreReloadTime time.Time
	lock          *sync.Mutex

	areaCodes map[string][][]string

	/**访问次数**/
	VisitNum int64
	/** 成功次数**/
	SuccessNum int64
	/** 失败次数**/
	FailedNum int64
}

func NewDisting(dataPath string, citydat string) (*Disting, error) {
	dist := &Disting{}
	dist.DataPath = dataPath
	dist.CodeDataPath = citydat
	dist.lock = new(sync.Mutex)
	return dist, nil
}

func (dist *Disting) LoadData() {

	var err error

	dist.DbCity, err = datx.NewCity(dist.DataPath)
	if err != nil {

		log.Println(err.Error())
		return
	}
	//加载城市编码
	err = dist.loadAreaCode(dist.CodeDataPath)
	if err != nil {

		log.Println(err.Error())
		return
	}

	finfo, err := os.Stat(dist.DataPath)
	if err != nil {

		log.Println(err.Error())
		return
	}

	dist.LatestModTime = finfo.ModTime().Unix()
	dist.PreReloadTime = time.Now()
}

//重新加载IP库
func (dist *Disting) ReloadData() {
	dist.lock.Lock()
	defer dist.lock.Unlock()

	finfo, err := os.Stat(dist.DataPath)

	db, err := datx.NewCity(dist.DataPath)
	if err != nil {

		log.Println("reload ipdat error:", err.Error())
		return
	}
	dist.DbCity = db

	dist.LatestModTime = finfo.ModTime().Unix()
	dist.PreReloadTime = time.Now()
	log.Println("reload ip ok")
}

//重新加载地区编码文件
func (dist *Disting) ReloadCityCode() {

	dist.loadAreaCode(dist.CodeDataPath)

	log.Println("reload code ok")
}

func (dist *Disting) FindDisting(ip string) *LocalData {

	dist.lock.Lock()
	defer dist.lock.Unlock()

	dist.VisitNum++

	data, err := dist.DbCity.Find(ip)

	if err != nil {
		log.Println("the ip %s can not find ", ip)
		return nil
	}

	//var codes []string
	code := dist.findCode(data[0], data[1], data[2])

	Info := &LocalData{}
	Info.Country = data[0]
	Info.Province = data[1]
	Info.CityName = data[2]
	if len(code) > 0 {
		Info.AreaCode = code[0]
		Info.CountryCode = code[15]
		Info.ProvinceCode = code[16]
	}

	dist.SuccessNum++
	return Info
}

//查找area code
func (dist *Disting) findCode(country string, province string, city string) []string {

	var res []string
	var data [][]string
	data = dist.areaCodes[country]
	if data == nil {
		return res
	}
	var name string
	var index int = 1

	//城市
	if city != "" {
		index = 7
		name = city
	}
	//省
	if name == "" && province != "" {
		index = 4
		name = province
	}
	//第2列,国家
	if name == "" && country != "" {
		index = 1
		name = country
	}

	for i := 0; i < len(data); i++ {
		if data[i][index] == name {
			log.Print(data[i])
			return data[i]
		}
	}
	//如果没有找到相应的只返回国家这一行，并把第一列致空
	if data[0] != nil {
		data[0][0] = ""
		log.Print(data[0])
		return data[0]
	}
	return res
}
func (dist *Disting) loadAreaCode(fn string) (err error) {

	//var err error

	var byteData []byte
	var file *os.File

	file, err = os.Open(fn)
	if err != nil {

		return err
	}
	byteData, err = ioutil.ReadAll(file)
	if err != nil {

		return err
	}
	strData := strings.Split(string(byteData), "\n")

	areaCodes := make(map[string][][]string, len(strData)-1)

	var country_code string
	var province_code string
	var pre_country string
	for i := 0; i < len(strData)-1; i++ {
		datas := strings.Split(string(strData[i]), "\t")
		if pre_country != "" && pre_country != datas[1] {
			country_code = ""
			province_code = ""
		}
		//国家编码
		if datas[4] == "" && datas[7] == "" {
			country_code = datas[0]
		}
		//省编码
		if datas[4] != "" && datas[7] == "" {
			province_code = datas[0]
		}
		datas = append(datas, country_code)
		datas = append(datas, province_code)

		country := datas[1]
		areaCodes[country] = append(areaCodes[country], datas)
		pre_country = datas[1]

	}
	dist.areaCodes = areaCodes

	return nil
}

func (dist *Disting) Watcher() {

	watcher, err := fsnotify.NewWatcher()

	if err != nil {

		log.Fatal(err.Error())
	}

	defer watcher.Close()
	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Print("event: ", event)
				dist.ReloadData()
				dist.ReloadCityCode()
			case err := <-watcher.Errors:
				log.Print(err)
			}
		}
	}()

	err = watcher.Add(dist.DataPath)
	if err != nil {

		log.Print(err.Error())
	}
	<-done
}
