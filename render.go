/**
* @program: stock
*
* @description:
*
* @author: lemo
*
* @create: 2022-06-01 19:25
**/

package main

import (
	"strings"
	"sync/atomic"
	"time"

	"github.com/jedib0t/go-pretty/text"
	"github.com/lemonyxk/charts"
	"github.com/lemonyxk/console"
)

func renderStockByCodeAndArea(area, code string) {
	isDataRun = true
	if mode == min {
		minRender(area, code)
	} else {
		dayRender(area, code)
	}
}

func dayRender(area, code string) {
	var index = 0

	var fn = func() {

		if index%20 == 0 {
			getDayData(area, code)
		}

		var realStr = renderRealData(realData([]string{area + code}))
		if len(realStr) == 0 {
			return
		}

		var l = charts.New([]string{
			time.Now().AddDate(0, 0, -365).Format("2006-01-02"),
			time.Now().AddDate(0, 0, -270).Format("2006-01-02"),
			time.Now().AddDate(0, 0, -180).Format("2006-01-02"),
			time.Now().AddDate(0, 0, -90).Format("2006-01-02"),
			time.Now().Format("2006-01-02"),
		}, priceData, 365)
		l.SetSize(termWidth-1, termHeight-3)
		l.SetYPrecision(2)
		l.RenderSymbol = func(lastValue float64, isLastEmpty bool, value float64, isEmpty bool, symbol string) string {
			if lastValue <= value {
				return console.FgRed.Sprint(symbol)
			}

			// return console.FgGreen.Sprint(symbol)
			return symbol
		}

		l.RenderXBorder = func(isEmpty bool, x string) string {
			if isEmpty {
				return "━"
			}
			return "┻"
		}
		// graph := asciigraph.Plot(
		// 	priceData,
		// 	asciigraph.Width(termWidth-8),
		// 	asciigraph.Height(termHeight-3),
		// 	// asciigraph.Caption(),
		// )

		flush()

		tips()

		write(l.Render())

		var s = strings.Repeat(" ", (termWidth-text.RuneCount(realStr[0]))/2)

		write("\n\n" + s + realStr[0])

		index++
	}

	fn()

	var ticker = time.NewTicker(time.Second * 3)
	atomic.AddInt32(&runningProcess, 1)
	go func() {

		for {
			select {
			case <-ticker.C:
				fn()
			case <-stopData:
				ticker.Stop()
				atomic.AddInt32(&runningProcess, -1)
				return
			}
		}
	}()
}

func minRender(area, code string) {
	var index = 0

	var fn = func() {

		if index%20 == 0 {
			getMinData(area, code)
		}

		var realStr = renderRealData(realData([]string{area + code}))
		if len(realStr) == 0 {
			return
		}

		var l = charts.New([]string{"09:30", "13:00", "15:00"}, priceData, 240)
		l.SetSize(termWidth-1, termHeight-3)
		l.SetYPrecision(2)
		l.RenderSymbol = func(lastValue float64, isLastEmpty bool, value float64, isEmpty bool, symbol string) string {
			if lastValue <= value {
				return console.FgRed.Sprint(symbol)
			}

			// return console.FgGreen.Sprint()
			return symbol
		}

		l.RenderEmpty = func(lastValue float64, isLastEmpty bool, value float64, isEmpty bool, empty string) string {
			return " "
		}

		l.RenderXBorder = func(isEmpty bool, x string) string {
			return "━"
		}

		// graph := asciigraph.Plot(
		// 	priceData,
		// 	asciigraph.Width(termWidth-8),
		// 	asciigraph.Height(termHeight-3),
		// 	// asciigraph.Caption(realStr),
		// )

		if isSelectMenu {
			return
		}

		flush()

		tips()

		write(l.Render())

		var s = strings.Repeat(" ", (termWidth-text.RuneCount(realStr[0]))/2)

		write("\n\n" + s + realStr[0])

		index++
	}

	go fn()

	var ticker = time.NewTicker(time.Second * 3)
	atomic.AddInt32(&runningProcess, 1)

	go func() {

		for {
			select {
			case <-ticker.C:
				go fn()
			case <-stopData:
				ticker.Stop()
				atomic.AddInt32(&runningProcess, -1)
				return
			}
		}
	}()

}
