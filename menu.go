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
	"strings"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/lemonyxk/console"
)

var Stdout = os.Stdout

var stopMenu = make(chan struct{})

func init() {

	go func() {

		if err := keyboard.Open(); err != nil {
			panic(err)
		}

		defer func() { _ = keyboard.Close() }()

		for {
			char, key, err := keyboard.GetKey()
			if err != nil {
				panic(err)
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
				case 'l':
					selectMenu()
				case 'd':
					changeModeToDay()
				case 'm':
					changeModeToMinute()
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

func menuTips() {
	var sm = mode
	if mode == day {
		sm += " 365"
	}
	var str = "[Mode: " + sm + "] [Q:Quit]\r\n"
	var s = strings.Repeat(" ", (termWidth-8-len(str))/2)
	write(console.FgYellow.Sprint(s + str))
}

func changeModeToDay() {
	if isSelectMenu {
		return
	}
	if mode == day {
		return
	}
	mode = day
	stop <- struct{}{}

	renderStockByCodeAndArea(menu[x-2].area, menu[x-2].code)
}

func changeModeToMinute() {
	if isSelectMenu {
		return
	}
	if mode == min {
		return
	}
	mode = min
	stop <- struct{}{}

	renderStockByCodeAndArea(menu[x-2].area, menu[x-2].code)
}

var isSelectMenu = false

func flush() {
	write("\x1bc")
	write("\033[H\033[3J")
}

func write(str string) {
	_, _ = fmt.Fprint(Stdout, str)
}

func exit() {
	if isSelectMenu {
		reset()
		return
	}
	flush()
	os.Exit(0)
}

type config struct {
	area    string
	code    string
	name    string
	change  string
	percent string
}

var x, y = 2, 1
var oldX, oldY = x, y

var menu = []config{
	{area: "sh", code: "000001", name: "上证指数"},
	{area: "sh", code: "000300", name: "沪深300"},
	{area: "sh", code: "600519", name: "贵州茅台"},
	{area: "sz", code: "399006", name: "创业扳指"},
	{area: "sh", code: "603103", name: "横店影视"},
	{area: "sz", code: "000651", name: "格力电器"},
	{area: "sh", code: "601318", name: "中国平安"},
}

func selectMenu() {
	if isSelectMenu {
		return
	}

	oldX, oldY = x, y
	isSelectMenu = true
	stop <- struct{}{}

	renderMenu()
}

func renderMenu() {

	var ticker = time.NewTicker(time.Second * 3)

	var fn = func(need bool) {
		var table = console.NewTable()
		table.Style().Options.DrawBorder = false
		table.Style().Options.SeparateColumns = false

		for i := 0; i < len(menu); i++ {
			var change = menu[i].change
			var percentStr = menu[i].percent
			if need {
				_, absoluteChange, percent := realData(menu[i].area, menu[i].code)
				if percent >= 0 {
					change = console.FgRed.Sprintf("+%s", absoluteChange)
					percentStr = console.FgRed.Sprintf("+%.2f%%", percent)
				} else {
					change = console.FgGreen.Sprintf("%s", absoluteChange)
					percentStr = console.FgGreen.Sprintf("%.2f%%", percent)
				}
				time.Sleep(time.Millisecond * 100)
				menu[i].change = change
				menu[i].percent = percentStr
			}
			table.Row(menu[i].name, menu[i].area, menu[i].code, change, percentStr)
			// write(fmt.Sprintf(" %s %s %s\n", menu[i].name, menu[i].area, menu[i].code))
		}

		if !isSelectMenu {
			return
		}

		flush()

		menuTips()

		write(table.Render())

		resetCursor()
	}

	fn(false)
	go fn(true)

	go func() {
		for {
			select {
			case <-ticker.C:
				go fn(true)
			case <-stopMenu:
				ticker.Stop()
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

func reset() {
	stopMenu <- struct{}{}
	isSelectMenu = false
	renderStockByCodeAndArea(menu[x-2].area, menu[x-2].code)
}

func enter() {
	stopMenu <- struct{}{}
	isSelectMenu = false
	x, y = oldX, oldY
	renderStockByCodeAndArea(menu[x-2].area, menu[x-2].code)
}

func left() {}

func right() {}

func up() {
	if oldX > 1+1 {
		oldX--
		resetCursor()
	}
}

func down() {
	if oldX < len(menu)+1 {
		oldX++
		resetCursor()
	}
}
