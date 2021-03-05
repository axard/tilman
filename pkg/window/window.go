package window

import (
	"fmt"

	"github.com/axard/tilman/pkg/clipregion"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ButtonAlignment int

const (
	ButtonAlignLeft  ButtonAlignment = tview.AlignLeft
	ButtonAlignRight ButtonAlignment = tview.AlignRight
)

// Button represents a button on the window title bar
type Button struct {
	Alignment ButtonAlignment
	Symbol    rune // icon for the button

	OnClick func(w *Window, b *Button) // callback to be invoked when the button is clicked

	offsetX, offsetY int
}

// Window defines a basic window
type Window struct {
	*tview.Box
	// The item contained in the window
	root tview.Primitive
	// window buttons on the title bar
	buttons []*Button
	// whether to render a border
	border bool
	// The color of the title.
	titleColor tcell.Color
	// The alignment of the title.
	titleAlign int
}

func New() *Window {
	window := &Window{
		Box:        tview.NewBox(),
		titleColor: tview.Styles.TitleColor,
		titleAlign: tview.AlignCenter,
	}

	window.Box.SetDrawFunc(window.drawBox)

	return window
}

// GetRoot returns the primitive that represents the main content of the window
func (w *Window) GetRoot() tview.Primitive {
	return w.root
}

// SetRoot sets the main content of the window
func (w *Window) SetRoot(root tview.Primitive) *Window {
	w.root = root
	return w
}

// Focus is called when this primitive receives focus.
func (w *Window) Focus(delegate func(p tview.Primitive)) {
	if w.root != nil {
		delegate(w.root)
	} else {
		delegate(w.Box)
	}
}

// HasFocus returns whether or not this primitive has focus.
func (w *Window) HasFocus() bool {
	if w.root != nil {
		return w.root.HasFocus()
	}
	return w.Box.HasFocus()
}

func (w *Window) Blur() {
	if w.root != nil {
		w.root.Blur()
	}
	w.Box.Blur()
}

// HasBorder returns true if this window has a border
// windows without border cannot be resized or dragged by the user
func (w *Window) HasBorder() bool {
	return w.border
}

// SetBorder sets the flag indicating whether or not the box should have a
// border.
func (w *Window) SetBorder(show bool) *Window {
	w.border = show
	w.Box.SetBorder(show)
	return w
}

// SetTitle sets the window title
func (w *Window) SetTitle(text string) *Window {
	w.Box.SetTitle(text)
	return w
}

// Draw draws this primitive on to the screen
func (w *Window) Draw(screen tcell.Screen) {
	if w.HasFocus() { // if the window has focus, make sure the underlying box shows a thicker border
		w.Box.Focus(nil)
	} else {
		w.Box.Blur()
	}
	w.Box.Draw(screen) // draw the window frame

	// draw the underlying root primitive within the window bounds
	if w.root != nil {
		x, y, width, height := w.GetInnerRect()
		w.root.SetRect(x, y, width, height)
		w.root.Draw(clipregion.New(screen, x, y, width, height))
	}

	// draw the window border
	if w.border {
		x, y, width, height := w.GetRect()
		screen = clipregion.New(screen, x, y, width, height)
		for _, button := range w.buttons {
			buttonX, buttonY := button.offsetX+x, button.offsetY+y
			if button.offsetX < 0 {
				buttonX += width
			}
			if button.offsetY < 0 {
				buttonY += height
			}

			// render the window title buttons
			tview.Print(screen, tview.Escape(fmt.Sprintf("[%c]", button.Symbol)), buttonX-1, buttonY, 9, 0, tcell.ColorYellow)
		}
	}
}

func (w *Window) drawBox(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
	def := tcell.StyleDefault

	// Fill background.
	background := def.Background(w.Box.GetBackgroundColor())
	if w.Box.GetBackgroundColor() != tcell.ColorDefault {
		for y_ := y; y_ < y+height; y_++ {
			for x_ := x; x_ < x+width; x_++ {
				screen.SetContent(x_, y_, ' ', nil, background)
			}
		}
	}

	if w.border && width >= 2 && height >= 2 {
		var horizontal rune
		if w.Box.HasFocus() {
			horizontal = tview.Borders.HorizontalFocus
		} else {
			horizontal = tview.Borders.Horizontal
		}
		borderStyle := tcell.StyleDefault.Foreground(tview.Styles.BorderColor)
		borderStyle = borderStyle.Attributes(w.Box.GetBorderAttributes()).Foreground(w.Box.GetBorderColor())
		for x_ := x; x_ < x+width; x_++ {
			screen.SetContent(x_, y, horizontal, nil, borderStyle)
		}

		// Draw title.
		title := w.Box.GetTitle()
		titleColor := w.titleColor
		titleAlign := w.titleAlign
		if title != "" && width >= 4 {
			printed, _ := tview.Print(screen, title, x+1, y, width-2, titleAlign, titleColor)
			if len(title)-printed > 0 && printed > 0 {
				_, _, style, _ := screen.GetContent(x+width-2, y)
				fg, _, _ := style.Decompose()
				tview.Print(screen, string(tview.SemigraphicsHorizontalEllipsis), x+width-2, y, 1, tview.AlignLeft, fg)
			}
		}
	}

	y += 1
	height -= 1

	return x, y, width, height
}

// AddButton adds a new window button to the title bar
func (w *Window) AddButton(symbol rune, alignment ButtonAlignment, onclick func(w *Window, b *Button)) *Window {
	w.buttons = append(w.buttons, &Button{
		Symbol:    symbol,
		Alignment: alignment,
		OnClick:   onclick,
	})

	offsetLeft, offsetRight := 2, -3
	for _, button := range w.buttons {
		if button.Alignment == ButtonAlignRight {
			button.offsetX = offsetRight
			offsetRight -= 3
		} else {
			button.offsetX = offsetLeft
			offsetLeft += 3
		}
	}

	return w
}

func (w *Window) RemoveButton(i int) *Window {
	if i < 0 || i >= len(w.buttons) {
		return w
	}

	w.buttons = append(w.buttons[:i], w.buttons[i+1:]...)

	return w
}

// GetButton returns the given button
func (w *Window) GetButton(i int) *Button {
	if i < 0 || i >= len(w.buttons) {
		return nil
	}

	return w.buttons[i]
}

// CountButtons returns the number of buttons in the window title bar
func (w *Window) CountButtons() int {
	return len(w.buttons)
}

func (w *Window) ClearButtons() *Window {
	w.buttons = nil
	return w
}

// MouseHandler returns a mouse handler for this primitive
func (w *Window) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return w.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if !w.InRect(event.Position()) {
			return false, nil
		}

		if action == tview.MouseLeftClick {
			x, y := event.Position()
			wx, wy, width, _ := w.GetRect()

			// check if any window button was pressed
			// if the window does not have border, it cannot receive button events
			if y == wy && w.border {
				for _, button := range w.buttons {
					if button.offsetX >= 0 && x == wx+button.offsetX || button.offsetX < 0 && x == wx+width+button.offsetX {
						if button.OnClick != nil {
							button.OnClick(w, button)
						}
						return true, nil
					}
				}
			}
		}
		// pass on clicks to the root primitive, if any
		if w.root != nil {
			return w.root.MouseHandler()(action, event, setFocus)
		}

		return w.Box.MouseHandler()(action, event, setFocus)
	})
}

// InputHandler returns a handler which receives key events when it has focus.
func (w *Window) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	if w.root != nil {
		return w.root.InputHandler()
	}
	return nil
}
