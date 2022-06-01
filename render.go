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
	"unicode/utf8"

	"github.com/guptarohit/asciigraph"
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

		graph := asciigraph.Plot(
			priceData,
			asciigraph.Width(termWidth-8),
			asciigraph.Height(termHeight-3),
			// asciigraph.Caption(),
		)

		flush()

		tips()

		write(graph)

		var s = strings.Repeat(" ", (termWidth-utf8.RuneCountInString(realStr[0]))/2)

		write("\n" + s + realStr[0])

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

		graph := asciigraph.Plot(
			priceData,
			asciigraph.Width(termWidth-8),
			asciigraph.Height(termHeight-3),
			// asciigraph.Caption(realStr),
		)

		if isSelectMenu {
			return
		}

		flush()

		tips()

		write(graph)

		var s = strings.Repeat(" ", (termWidth-utf8.RuneCountInString(realStr[0]))/2)

		write("\n" + s + realStr[0])

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
