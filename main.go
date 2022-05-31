/**
* @program: stock
*
* @description:
*
* @author: lemo
*
* @create: 2022-05-31 14:39
**/
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/guptarohit/asciigraph"
	jsoniter "github.com/json-iterator/go"
	"github.com/lemonyxk/console"
	"github.com/lemonyxk/kitty/v2/socket/http/client"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

// 分钟
var minURL = `https://web.ifzq.gtimg.cn/appstock/app/minute/query?code=`

// 日线
// https://web.ifzq.gtimg.cn/appstock/app/fqkline/get?param=sh000001,day,,,320,hfq

// 实时
var realURL = `https://web.sqt.gtimg.cn/q=`

var code = ""

var area = ""

func GetFlagAndArgs(flag []string, args []string) (string, string) {
	for i := 0; i < len(args); i++ {
		for j := 0; j < len(flag); j++ {
			if args[i] == flag[j] {
				if i+1 < len(args) {
					return flag[j], args[i+1]
				}
			}
		}
	}
	return "", ""
}

func main() {

	_, code = GetFlagAndArgs([]string{"code", "--code", "-c"}, os.Args[1:])
	if code == "" {
		code = "000001"
	}

	_, area = GetFlagAndArgs([]string{"area", "--area", "-a"}, os.Args[1:])
	if area == "" {
		area = "sh"
	}

	var index = 0

	getData()

	termWidth, termHeight, err := terminal.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}

	go func() {
		for {

			if index%60 == 0 {
				getData()
			}

			graph := asciigraph.Plot(
				priceData,
				asciigraph.Width(termWidth-8),
				asciigraph.Height(termHeight-4),
				asciigraph.Caption(realData()),
			)

			fmt.Print("\x1bc")

			fmt.Print("\033[H\033[3J")

			fmt.Print("\r\n")

			fmt.Println(graph)

			index++
			time.Sleep(time.Second * 3)
		}
	}()

	select {}

}

// http get method

var timeData []float64
var priceData []float64

func getData() {

	timeData = timeData[:0]
	priceData = priceData[:0]

	// var m = make(map[float64]float64)

	var res = client.Get(minURL + area + code).Query().Send()

	var arrStr = jsoniter.Get(res.Bytes(), "data", area+code, "data", "data").ToString()

	var arr []string

	_ = jsoniter.Unmarshal([]byte(arrStr), &arr)

	for _, v := range arr {
		var a = strings.Split(v, " ")
		timeData = append(timeData, StringToFloat(a[0]))
		priceData = append(priceData, StringToFloat(a[1]))
	}

}

func StringToFloat(str string) float64 {
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0
	}
	return f
}

func realData() string {

	var now = time.Now().Format("2006-01-02 15:04:05")

	var res = client.Get(realURL + area + code).Query().Send()
	var bts, _ = GbkToUtf8(res.Bytes())
	var str = string(bts)

	var arr = strings.Split(str, "=")
	var data = arr[1]
	if len(data) < 2 {
		return ""
	}

	data = strings.Replace(data, "\"", "", -1)
	data = strings.Replace(data, ";", "", -1)
	var dataArr = strings.Split(data, "~")
	if len(dataArr) < 6 {
		return ""
	}

	var title = dataArr[1]
	var code = dataArr[2]
	var currentPrice = dataArr[3]
	var lowestPrice = dataArr[4]
	var openPrice = dataArr[5]

	var percent = (StringToFloat(currentPrice) - StringToFloat(openPrice)) / StringToFloat(openPrice) * 100

	var percentStr string

	if percent > 0 {
		percentStr = console.FgRed.Sprintf("%.2f%%", percent)
	} else {
		percentStr = console.FgGreen.Sprintf("%.2f%%", percent)
	}

	var ts = fmt.Sprintf("%s %s current: %s ( %s ) lowest: %s open: %s time: %s", title, code, currentPrice, percentStr, lowestPrice, openPrice, now)

	return ts
}
