package main

import (
	"disting"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var dist *disting.Disting

func getIpDisting(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	w.Header().Set("Content-Type", "text/plain;charset=utf-8")
	strIp := r.Form["ip"]

	if len(strIp) == 0 || len(strIp[0]) == 0 {
		fmt.Fprintf(w, "{}")
		return
	}

	distinfo := dist.FindDisting(strIp[0])
	if distinfo == nil {

		fmt.Fprintf(w, "{}")
		return
	}

	jsonData, _ := json.Marshal(distinfo)

	fmt.Fprintf(w, string(jsonData))
}

func reloadData(w http.ResponseWriter, r *http.Request) {

	dist.ReloadData()
	w.Header().Set("Content-Type", "text/plain;charset=utf-8")
	fmt.Fprintf(w, "ok")
}

func statusInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html;charset=utf-8")
	format := `
	<html>
	<body>
		ipdata更新时间: %s <br \>
		数据加载时间: %s <br \>
		Visit :%d <br \>
		Success:%d <br \>
		Failed:%d <br \>
		SuccessRate:%.2f
		</body>
		</html>
	`
	rate := 0.0
	if dist.VisitNum > 0 && dist.SuccessNum > 0 {
		rate = float64(dist.SuccessNum / dist.VisitNum)
	}
	tm := time.Unix(dist.LatestModTime, 0).Format("2006-01-02 15:04:05")
	fmt.Fprintf(w, format, tm, dist.PreReloadTime, dist.VisitNum, dist.SuccessNum, dist.FailedNum, rate*100)
}

func main() {

	port := flag.String("port", ":9873", "ip server port")
	ipdatx := flag.String("ipdat", "", "ip data path")
	citydat := flag.String("codedat", "", "city data path")

	flag.Parse()

	_, err := os.Stat(*ipdatx)
	if err != nil {

		fmt.Print("Can not load the ip data in [%s]", *ipdatx)
		return
	}
	if _, err := os.Stat(*citydat); err != nil {

		fmt.Print("Can not load the city code data in [%s]", *citydat)
		return
	}

	dist, err = disting.NewDisting(*ipdatx, *citydat)

	//加载ip数据文件
	dist.LoadData()

	//启动ip库文件监听
	go dist.Watcher()

	http.HandleFunc("/lbs", getIpDisting)
	http.HandleFunc("/status", statusInfo)

	log.Println("http listen in port:", *port)

	err = http.ListenAndServe(*port, nil)

	if err != nil {

		log.Fatal(err.Error())
	}
}
