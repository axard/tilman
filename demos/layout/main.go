package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/axard/tilman"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const corporate = `Leverage agile frameworks to provide a robust synopsis for high level overviews. Iterative approaches to corporate strategy foster collaborative thinking to further the overall value proposition. Organically grow the holistic world view of disruptive innovation via workplace diversity and empowerment.

Bring to the table win-win survival strategies to ensure proactive domination. At the end of the day, going forward, a new normal that has evolved from generation X is on the runway heading towards a streamlined cloud solution. User generated content in real-time will have multiple touchpoints for offshoring.

Capitalize on low hanging fruit to identify a ballpark value added activity to beta test. Override the digital divide with additional clickthroughs from DevOps. Nanotechnology immersion along the information highway will close the loop on focusing solely on the bottom line.

[yellow]Press Enter, then Tab/Backtab for word selections`

var App = tview.NewApplication()

func textview() tview.Primitive {
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetChangedFunc(func() {
			App.Draw()
		})

	textView.SetBorder(false)

	numSelections := 0
	go func() {
		for _, word := range strings.Split(corporate, " ") {
			if word == "the" {
				word = "[#ff0000]the[white]"
			}
			if word == "to" {
				word = fmt.Sprintf(`["%d"]to[""]`, numSelections)
				numSelections++
			}
			fmt.Fprintf(textView, "%s ", word)
			time.Sleep(50 * time.Millisecond)
		}
	}()

	textView.SetDoneFunc(func(key tcell.Key) {
		currentSelection := textView.GetHighlights()
		if key == tcell.KeyEnter {
			if len(currentSelection) > 0 {
				textView.Highlight()
			} else {
				textView.Highlight("0").ScrollToHighlight()
			}
		} else if len(currentSelection) > 0 {
			index, _ := strconv.Atoi(currentSelection[0])
			if key == tcell.KeyTab {
				index = (index + 1) % numSelections
			} else if key == tcell.KeyBacktab {
				index = (index - 1 + numSelections) % numSelections
			} else {
				return
			}
			textView.Highlight(strconv.Itoa(index)).ScrollToHighlight()
		}
	})

	return textView
}

func main() {
	window_1 := tilman.NewWindow().
		SetRoot(textview()).
		SetTitle("Agile-1").
		SetBorder(true).
		AddButton('X', tilman.WindowButtonAlignRight, func(_ *tilman.Window, _ *tilman.WindowButton) {
			App.Stop()
		})

	window_2 := tilman.NewWindow().
		SetRoot(textview()).
		SetTitle("Agile-2").
		SetBorder(true).
		AddButton('X', tilman.WindowButtonAlignRight, func(_ *tilman.Window, _ *tilman.WindowButton) {
			App.Stop()
		})

	window_3 := tilman.NewWindow().
		SetRoot(textview()).
		SetTitle("Agile-3").
		SetBorder(true).
		AddButton('X', tilman.WindowButtonAlignRight, func(_ *tilman.Window, _ *tilman.WindowButton) {
			App.Stop()
		})

	layout_1 := tilman.NewLayout().
		SetDirection(tilman.HorizontalLayout).
		SetSplitter(true).
		AddItem(window_1, tilman.AutoSize).
		AddItem(window_2, tilman.AutoSize)

	layout_2 := tilman.NewLayout().
		SetDirection(tilman.VerticalLayout).
		SetSplitter(true).
		AddItem(layout_1, tilman.AutoSize).
		AddItem(window_3, tilman.AutoSize)

	if err := App.SetRoot(layout_2, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
