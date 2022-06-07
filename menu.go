/**
* @program: stock
*
* @description:
*
* @author: lemo
*
* @create: 2022-05-31 21:48
**/

package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/lemonyxk/console"
	"github.com/lemonyxk/utils/v3"
)

var Stdout = os.Stdout

var stopMenu = make(chan struct{})

func init() {
	start()
}

func start() {

	go func() {

		if err := keyboard.Open(); err != nil {
			panic(err)
		}

		defer func() { _ = keyboard.Close() }()

		for {
			char, key, err := keyboard.GetKey()
			if err != nil {
				exit()
			}

			switch key {
			case 0xffea:
				ctrlC(key)
			case 0xffeb:
				ctrlC(key)
			case 0xffec:
				ctrlC(key)
			case 0xffed:
				ctrlC(key)
			case 0x000d:
				ctrlC(key)
			default:
				switch char {
				// case 'l':
				// 	selectMenu()
				case 'd':
					changeModeToDay()
				case 'm':
					changeModeToMinute()
				case 'e':
					go editMenu()
					return
				case 'b':
					backMenu()
				case 'q':
					exit()
				default:
					// exit()
				}
			}

			time.Sleep(time.Millisecond * 50)
		}
	}()
}

func editMenu() {
	if !isSelectMenu {
		return
	}

	showCursor()

	stopMenu <- struct{}{}

	flush()

	isEditMenu = true

	editFile(configPath)

	var res = utils.File.ReadFromPath(configPath)
	_ = utils.Json.Decode(res.Bytes(), &menu)

	isEditMenu = false

	start()

	renderMenu()
}

func changeModeToDay() {
	if isSelectMenu {
		return
	}
	if mode == day {
		return
	}

	mode = day

	if isDataRun {
		stopData <- struct{}{}
		isDataRun = false
	}
	hideCursor()
	renderStockByCodeAndArea(menu[x-fixX].Area, menu[x-fixX].Code)
}

func changeModeToMinute() {
	if isSelectMenu {
		return
	}
	if mode == min {
		return
	}

	mode = min

	if isDataRun {
		stopData <- struct{}{}
		isDataRun = false
	}
	hideCursor()
	renderStockByCodeAndArea(menu[x-fixX].Area, menu[x-fixX].Code)
}

var isSelectMenu = false
var isEditMenu = false

func flush() {
	write("\033[H\033[J")
}

func write(str string) {
	_, _ = fmt.Fprint(Stdout, str)
}

func backMenu() {
	if isSelectMenu {
		return
	}

	if isDataRun {
		back()
	}
}

func exit() {
	flush()
	showCursor()
	_ = keyboard.Close()
	os.Exit(0)
}

func showCursor() {
	write("\033[?25h")
}

func hideCursor() {
	write("\033[?25l")
}

type config struct {
	Area string
	Code string
	Name string
	// Change  string
	// Percent string
}

var fixX = 3

var x, y = fixX, 1
var oldX, oldY = x, y

var menu []config

// {area: "sh", code: "000001", name: "上证指数"},
// {area: "sh", code: "000300", name: "沪深300"},
// {area: "sh", code: "600519", name: "贵州茅台"},
// {area: "sh", code: "603103", name: "横店影视"},
// {area: "sh", code: "601318", name: "中国平安"},
// {area: "sz", code: "399006", name: "创业扳指"},
// {area: "sz", code: "000651", name: "格力电器"},

func selectMenu() {
	if isSelectMenu {
		return
	}

	showCursor()

	oldX, oldY = x, y
	isSelectMenu = true

	if isDataRun {
		stopData <- struct{}{}
		isDataRun = false
	}

	renderMenu()
}

type show struct {
	config
	Change       string
	Percent      string
	Current      string
	HighestPrice string
	LowestPrice  string
}

func renderMenu() {

	var fn = func(need bool) {
		var table = console.NewTable()
		table.Style().Options.DrawBorder = false
		table.Style().Options.SeparateColumns = false

		var showStr []show
		var params []string
		for i := 0; i < len(menu); i++ {
			showStr = append(showStr, show{config: menu[i]})
			params = append(params, menu[i].Area+menu[i].Code)
		}

		var data = realData(params)
		for i := 0; i < len(data); i++ {
			if !need {
				showStr[i].Change = console.FgRed.Sprintf("+%s", "0")
				showStr[i].Percent = console.FgRed.Sprintf("+%.2f%%", 0.00)
				showStr[i].Current = console.FgRed.Sprintf("↑ %s", "0")
				showStr[i].HighestPrice = console.FgHiRed.Sprintf("%s", "0")
				showStr[i].LowestPrice = console.FgHiGreen.Sprintf("%s", "0")
				continue
			}

			var percentStr = showStr[i].Percent
			var change = showStr[i].Change
			var absoluteChange = data[i][31]
			var percent = StringToFloat(data[i][32])
			var currentPrice = data[i][3]
			var highestPrice = data[i][33]
			var lowestPrice = data[i][34]
			if percent >= 0 {
				currentPrice = console.FgRed.Sprintf("↑ %s", currentPrice)
				change = console.FgRed.Sprintf("+%s", absoluteChange)
				percentStr = console.FgRed.Sprintf("+%.2f%%", percent)
			} else {
				currentPrice = console.FgGreen.Sprintf("↓ %s", currentPrice)
				change = console.FgGreen.Sprintf("%s", absoluteChange)
				percentStr = console.FgGreen.Sprintf("%.2f%%", percent)
			}
			highestPrice = console.FgHiRed.Sprintf("%s", highestPrice)
			lowestPrice = console.FgHiGreen.Sprintf("%s", lowestPrice)

			showStr[i].Change = change
			showStr[i].Percent = percentStr
			showStr[i].Current = currentPrice
			showStr[i].HighestPrice = highestPrice
			showStr[i].LowestPrice = lowestPrice
		}

		for i := 0; i < len(menu); i++ {
			table.Row(
				showStr[i].Name, showStr[i].Area,
				showStr[i].Code, showStr[i].Current,
				showStr[i].Percent, showStr[i].Change,
				showStr[i].LowestPrice, showStr[i].HighestPrice,
			)
		}

		if !isSelectMenu {
			return
		}

		if isEditMenu {
			return
		}

		flush()

		menuTips()

		write(table.Render())

		resetCursor()
	}

	fn(false)
	go fn(true)

	var ticker = time.NewTicker(time.Second * 3)
	atomic.AddInt32(&runningProcess, 1)

	go func() {
		for {
			select {
			case <-ticker.C:
				go fn(true)
			case <-stopMenu:
				ticker.Stop()
				atomic.AddInt32(&runningProcess, -1)
				return
			}
		}
	}()
}

func ctrlC(key keyboard.Key) {
	if !isSelectMenu {
		return
	}

	switch key {
	case 0xffea:
		// right()
	case 0xffeb:
		// left()
	case 0xffec:
		down()
	case 0xffed:
		up()
	case 0x000d:
		enter()
	}
}

func resetCursor() {
	write(fmt.Sprintf("\033[%d;%dH", oldX, oldY))
}

func back() {
	stopData <- struct{}{}
	isDataRun = false
	selectMenu()
}

func enter() {
	stopMenu <- struct{}{}
	isSelectMenu = false
	x, y = oldX, oldY
	hideCursor()
	renderStockByCodeAndArea(menu[x-fixX].Area, menu[x-fixX].Code)
}

func left() {}

func right() {}

func up() {
	if oldX > fixX {
		oldX--
		resetCursor()
	}
}

func down() {
	if oldX < len(menu)+fixX-1 {
		oldX++
		resetCursor()
	}
}

// edit mode like git
func editFile(file string) {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "notepad "+file)
	} else {
		cmd = exec.Command(os.Getenv("SHELL"), "-c", "vim "+file)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}
