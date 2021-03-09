package tilman

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Direction int

const (
	VerticalLayout Direction = iota
	HorizontalLayout
)

const (
	AutoSize = 0
)

type Item struct {
	tview.Primitive

	Size int
}

type splitter struct {
	x, y [2]int // begin and end points

	a, b *Item
}

func (s *splitter) contain(x, y int) bool {
	if s.x[0] == s.x[1] {
		// vertical splitter on horizontal direction
		return s.x[0] == x && (s.y[0] <= y && y <= s.y[1])
	} else {
		// horizontal splitter on vertical direction
		return s.y[0] == y && (s.x[0] <= x && x <= s.x[1])
	}
}

type Layout struct {
	// The items contained in the layout
	items []*Item

	dragX, dragY int

	focusedSplitterNumber int
	draggedSplitter       *splitter
	splitters             []*splitter

	// Whether or not a splitterFlag is drawn, reducing the box's space for content by
	// two in width and height.
	splitterFlag bool

	// The border style.
	splitterStyle tcell.Style

	direction Direction

	x, y, width, height int

	// The layout's background color.
	backgroundColor tcell.Color

	// An optional capture function which receives a key event and returns the
	// event to be forwarded to the primitive's default input handler (nil if
	// nothing should be forwarded).
	inputCapture func(event *tcell.EventKey) *tcell.EventKey

	// An optional capture function which receives a mouse event and returns the
	// event to be forwarded to the primitive's default mouse event handler (at
	// least one nil if nothing should be forwarded).
	mouseCapture func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse)
}

func NewLayout() *Layout {
	layout := &Layout{
		backgroundColor:       tview.Styles.PrimitiveBackgroundColor,
		focusedSplitterNumber: -1,
		splitterStyle:         tcell.StyleDefault.Foreground(tview.Styles.BorderColor),
	}

	return layout
}

func (l *Layout) SetDirection(d Direction) *Layout {
	l.direction = d
	return l
}

func (l *Layout) GetDirection() Direction {
	return l.direction
}

// SetBackgroundColor sets the layout's background color.
func (l *Layout) SetBackgroundColor(color tcell.Color) *Layout {
	l.backgroundColor = color
	return l
}

// SetSplitter sets the flag indicating whether or not the layout should render a
// splitters between primitives
func (l *Layout) SetSplitter(show bool) *Layout {
	l.splitterFlag = show
	return l
}

// SetSplitterColor sets the layout's splitter color.
func (l *Layout) SetSplitterColor(color tcell.Color) *Layout {
	l.splitterStyle = l.splitterStyle.Foreground(color)
	return l
}

// SetSplitterAttributes sets the splitter's style attributes. You can combine
// different attributes using bitmask operations:
//
//   layout.SetSplitterAttributes(tcell.AttrUnderline | tcell.AttrBold)
func (l *Layout) SetSplitterAttributes(attr tcell.AttrMask) *Layout {
	l.splitterStyle = l.splitterStyle.Attributes(attr)
	return l
}

// GetSplitterAttributes returns the splitter's style attributes.
func (l *Layout) GetSplitterAttributes() tcell.AttrMask {
	_, _, attr := l.splitterStyle.Decompose()
	return attr
}

// GetSplitterColor returns the layout's splitter color.
func (l *Layout) GetSplitterColor() tcell.Color {
	color, _, _ := l.splitterStyle.Decompose()
	return color
}

// GetBackgroundColor returns the layout's background color.
func (l *Layout) GetBackgroundColor() tcell.Color {
	return l.backgroundColor
}

func (l *Layout) itemsAmount() (int, int) {
	auto := 0
	for _, item := range l.items {
		if item.Size == AutoSize {
			auto += 1
		}
	}

	return len(l.items) - auto, auto
}

func (l *Layout) itemsSize() int {
	size := 0
	for _, item := range l.items {
		size += item.Size
	}

	return size
}

func (l *Layout) availableSpace() int {
	switch l.direction {
	case HorizontalLayout:
		return l.width
	case VerticalLayout:
		return l.height
	default:
		return 0
	}
}

func (l *Layout) splittersAmount() int {
	if len(l.items) <= 1 {
		return 0
	} else {
		return len(l.items) - 1
	}
}

func (l *Layout) Draw(screen tcell.Screen) {
	x, y, width, height := l.GetRect()
	def := tcell.StyleDefault

	// Fill background.
	background := def.Background(l.backgroundColor)
	if l.backgroundColor != tcell.ColorDefault {
		for y_ := y; y_ < y+height; y_++ {
			for x_ := x; x_ < x+width; x_++ {
				screen.SetContent(x_, y_, ' ', nil, background)
			}
		}
	}

	_, auto := l.itemsAmount()
	size := l.itemsSize()
	space := l.availableSpace()
	seps := l.splittersAmount()

	autoSize := 0
	if auto != 0 {
		autoSize = (space - size - seps) / auto
	}

	switch l.direction {
	case HorizontalLayout:
		vertical := tview.Borders.Vertical

		for number, item := range l.items {
			if item.Size == AutoSize {
				item.Primitive.SetRect(x, y, autoSize, height)
				item.Primitive.Draw(NewClipRegion(screen, x, y, autoSize, height))
				x += autoSize
			} else {
				item.Primitive.SetRect(x, y, item.Size, height)
				item.Primitive.Draw(NewClipRegion(screen, x, y, item.Size, height))
				x += item.Size
			}

			if seps > 0 {
				if l.splitterFlag {
					if number == l.focusedSplitterNumber {
						vertical = tview.Borders.VerticalFocus
					}

					for y_ := y; y_ < y+height; y_++ {
						screen.SetContent(x, y_, vertical, nil, l.splitterStyle)
					}
				}

				seps -= 1
				x += 1
			}
		}

	case VerticalLayout:
		horizontal := tview.Borders.Horizontal

		for number, item := range l.items {
			if item.Size == AutoSize {
				item.Primitive.SetRect(x, y, width, autoSize)
				item.Primitive.Draw(NewClipRegion(screen, x, y, width, autoSize))
				y += autoSize
			} else {
				item.Primitive.SetRect(x, y, width, item.Size)
				item.Primitive.Draw(NewClipRegion(screen, x, y, width, item.Size))
				y += item.Size
			}

			if seps > 0 {
				if l.splitterFlag {
					if number == l.focusedSplitterNumber {
						horizontal = tview.Borders.HorizontalFocus
					}

					for x_ := x; x_ < x+width; x_++ {
						screen.SetContent(x_, y, horizontal, nil, l.splitterStyle)
					}
				}

				seps -= 1
				y += 1
			}
		}
	}
}

func (l *Layout) GetRect() (int, int, int, int) {
	return l.x, l.y, l.width, l.height
}

func (l *Layout) SetRect(x, y, width, height int) {
	l.x = x
	l.y = y
	l.width = width
	l.height = height

	l.rebuildSplitters()
}

func (l *Layout) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return l.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		for _, item := range l.items {
			if item.Primitive.HasFocus() {
				if handler := item.Primitive.InputHandler(); handler != nil {
					handler(event, setFocus)
					return
				}
			}
		}
	})
}

// WrapInputHandler wraps an input handler (see InputHandler()) with the
// functionality to capture input (see SetInputCapture()) before passing it
// on to the provided (default) input handler.
//
// This is only meant to be used by subclassing primitives.
func (l *Layout) WrapInputHandler(inputHandler func(*tcell.EventKey, func(p tview.Primitive))) func(*tcell.EventKey, func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if l.inputCapture != nil {
			event = l.inputCapture(event)
		}
		if event != nil && inputHandler != nil {
			inputHandler(event, setFocus)
		}
	}
}

func (l *Layout) Focus(delegate func(p tview.Primitive)) {
	if l.focusedSplitterNumber == -1 {
		for _, item := range l.items {
			if item.Primitive != nil {
				delegate(item.Primitive)
				return
			}
		}
	}
}

func (l *Layout) HasFocus() bool {
	for _, item := range l.items {
		if item.Primitive != nil && item.Primitive.HasFocus() {
			return true
		}
	}
	return false
}

func (l *Layout) Blur() {
	l.focusedSplitterNumber = -1
	for _, item := range l.items {
		if item.Primitive != nil && item.Primitive.HasFocus() {
			item.Primitive.Blur()
		}
	}
}

// InRect returns true if the given coordinate is within the bounds of the box's
// rectangle.
func (l *Layout) InRect(x, y int) bool {
	rectX, rectY, width, height := l.GetRect()
	return x >= rectX && x < rectX+width && y >= rectY && y < rectY+height
}

// MouseHandler returns the mouse handler for this primitive.
func (l *Layout) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return l.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if !l.InRect(event.Position()) {
			return false, nil
		}

		// Pass mouse events along to the first child item that takes it.
		for _, item := range l.items {
			if item.Primitive != nil {
				consumed, capture = item.Primitive.MouseHandler()(action, event, setFocus)
				if consumed {
					l.focusedSplitterNumber = -1
					return
				}
			}
		}

		switch action {
		case tview.MouseMove:
			if l.draggedSplitter != nil {
				x, y := event.Position()
				dx, dy := x-l.dragX, y-l.dragY
				wxa, wya, wwa, wha := l.draggedSplitter.a.GetRect()
				wxb, wyb, wwb, whb := l.draggedSplitter.b.GetRect()

				l.dragX = x
				l.dragY = y

				switch l.direction {
				case HorizontalLayout:
					l.draggedSplitter.a.SetRect(wxa, wya, wwa+dx, wha)
					l.draggedSplitter.b.SetRect(wxb+dx, wyb, wwb-dx, whb)
					l.draggedSplitter.a.Size = wwa + dx
					l.draggedSplitter.b.Size = wwb - dx
				case VerticalLayout:
					l.draggedSplitter.a.SetRect(wxa, wya, wwa, wha+dy)
					l.draggedSplitter.b.SetRect(wxb, wyb+dy, wwb, whb-dy)
					l.draggedSplitter.a.Size = wha + dy
					l.draggedSplitter.b.Size = whb - dy
				default:
					panic(fmt.Sprintf("invalid layout direction: %v", l.direction))
				}
				return true, nil
			}

		case tview.MouseLeftDown:
			for number, splitter := range l.splitters {
				x, y := event.Position()
				if splitter.contain(x, y) {
					l.focusedSplitterNumber = number
					l.draggedSplitter = splitter
					l.dragX = x
					l.dragY = y

					for _, item := range l.items {
						if item.Primitive != nil {
							item.Primitive.Blur()
						}
					}

					return true, nil
				}
			}

		case tview.MouseLeftUp:
			l.draggedSplitter = nil
			l.rebuildSplitters()
			return true, nil
		}

		return
	})
}

// WrapMouseHandler wraps a mouse event handler (see MouseHandler()) with the
// functionality to capture mouse events (see SetMouseCapture()) before passing
// them on to the provided (default) event handler.
//
// This is only meant to be used by subclassing primitives.
func (l *Layout) WrapMouseHandler(mouseHandler func(tview.MouseAction, *tcell.EventMouse, func(p tview.Primitive)) (bool, tview.Primitive)) func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if l.mouseCapture != nil {
			action, event = l.mouseCapture(action, event)
		}
		if event != nil && mouseHandler != nil {
			consumed, capture = mouseHandler(action, event, setFocus)
		}
		return
	}
}

func (l *Layout) AddItem(p tview.Primitive, size int) *Layout {
	l.items = append(l.items, &Item{
		Primitive: p,
		Size:      size,
	})

	l.draggedSplitter = nil
	l.rebuildSplitters()

	return l
}

func (l *Layout) RemoveItem(i int) *Layout {
	if i < 0 || i >= len(l.items) {
		return l
	}

	l.draggedSplitter = nil
	l.items = append(l.items[:i], l.items[i+1:]...)
	l.rebuildSplitters()

	return l
}

func (l *Layout) GetItem(i int) *Item {
	if i < 0 || i >= len(l.items) {
		return nil
	}

	return l.items[i]
}

func (l *Layout) CountItems() int {
	return len(l.items)
}

func (l *Layout) ClearItems() *Layout {
	l.items = nil
	l.splitters = nil
	return l
}

// SetInputCapture installs a function which captures key events before they are
// forwarded to the primitive's default key event handler. This function can
// then choose to forward that key event (or a different one) to the default
// handler by returning it. If nil is returned, the default handler will not
// be called.
//
// Providing a nil handler will remove a previously existing handler.
//
// Note that this function will not have an effect on primitives composed of
// other primitives, such as Form, Flex, or Grid. Key events are only captured
// by the primitives that have focus (e.g. InputField) and only one primitive
// can have focus at a time. Composing primitives such as Form pass the focus on
// to their contained primitives and thus never receive any key events
// themselves. Therefore, they cannot intercept key events.
func (l *Layout) SetInputCapture(capture func(event *tcell.EventKey) *tcell.EventKey) *Layout {
	l.inputCapture = capture
	return l
}

// GetInputCapture returns the function installed with SetInputCapture() or nil
// if no such function has been installed.
func (l *Layout) GetInputCapture() func(event *tcell.EventKey) *tcell.EventKey {
	return l.inputCapture
}

// SetMouseCapture sets a function which captures mouse events (consisting of
// the original tcell mouse event and the semantic mouse action) before they are
// forwarded to the primitive's default mouse event handler. This function can
// then choose to forward that event (or a different one) by returning it or
// returning a nil mouse event, in which case the default handler will not be
// called.
//
// Providing a nil handler will remove a previously existing handler.
func (l *Layout) SetMouseCapture(capture func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse)) *Layout {
	l.mouseCapture = capture
	return l
}

// GetMouseCapture returns the function installed with SetMouseCapture() or nil
// if no such function has been installed.
func (l *Layout) GetMouseCapture() func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	return l.mouseCapture
}

func (l *Layout) rebuildSplitters() {
	l.splitters = nil

	x, y, width, height := l.GetRect()

	_, auto := l.itemsAmount()
	size := l.itemsSize()
	space := l.availableSpace()
	seps := l.splittersAmount()

	autoSize := 0
	if auto != 0 {
		autoSize = (space - size - seps) / auto
	}

	switch l.direction {
	case HorizontalLayout:
		for i := 0; i < len(l.items)-1; i++ {
			if l.items[i].Size == AutoSize {
				x += autoSize
			} else {
				x += l.items[i].Size
			}

			if seps > 0 {
				l.splitters = append(l.splitters, &splitter{
					x: [2]int{x, x},
					y: [2]int{y, y + height - 1},
					a: l.items[i],
					b: l.items[i+1],
				})

				seps -= 1
				x += 1
			}
		}

	case VerticalLayout:
		for i := 0; i < len(l.items)-1; i++ {
			if l.items[i].Size == AutoSize {
				y += autoSize
			} else {
				y += l.items[i].Size
			}

			if seps > 0 {
				l.splitters = append(l.splitters, &splitter{
					x: [2]int{x, x + width - 1},
					y: [2]int{y, y},
					a: l.items[i],
					b: l.items[i+1],
				})

				seps -= 1
				y += 1
			}
		}
	}
}
