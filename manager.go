package tilman

import (
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Manager struct {
	*tview.Box
	sync.Mutex

	logicalRoot tview.Primitive
	visibleRoot tview.Primitive

	restoreX, restoreY, restoreWidth, restoreHeight int
}

func NewWindowManager() *Manager {
	manager := &Manager{
		Box: tview.NewBox(),
	}

	return manager
}

func (m *Manager) SetRoot(root *Layout) *Manager {
	m.logicalRoot = root
	m.visibleRoot = root

	return m
}

func (m *Manager) GetRoot() *Layout {
	return m.logicalRoot.(*Layout)
}

func (m *Manager) Maximize(w *Window) *Manager {
	m.visibleRoot = w
	m.restoreX, m.restoreY, m.restoreWidth, m.restoreHeight = w.GetRect()

	return m
}

func (m *Manager) IsMaximazed(w *Window) bool {
	vw, ok := m.visibleRoot.(*Window)
	if !ok {
		return false
	}

	return vw == w
}

func (m *Manager) Restore() *Manager {
	m.visibleRoot.SetRect(m.restoreX, m.restoreY, m.restoreWidth, m.restoreHeight)
	m.visibleRoot = m.logicalRoot

	return m
}

// Focus is called when this primitive receives focus
func (m *Manager) Focus(delegate func(p tview.Primitive)) {
	m.Lock()
	defer m.Unlock()

	m.visibleRoot.Focus(delegate)
}

// HasFocus returns whether or not this primitive has focus.
func (m *Manager) HasFocus() bool {
	return m.visibleRoot.HasFocus()
}

// Draw draws this primitive onto the screen.
func (m *Manager) Draw(screen tcell.Screen) {
	m.Lock()
	defer m.Unlock()

	m.Box.Draw(screen)

	x, y, width, height := m.Box.GetInnerRect()
	m.visibleRoot.SetRect(x, y, width, height)
	m.visibleRoot.Draw(NewClipRegion(screen, x, y, width, height))
}

// MouseHandler returns the mouse handler for this primitive.
func (m *Manager) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return m.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		m.Lock()
		defer m.Unlock()

		// ignore mouse events out of the bounds of the window manager
		if !m.InRect(event.Position()) {
			return false, nil
		}

		return m.visibleRoot.MouseHandler()(action, event, setFocus)
	})
}

// InputHandler returns a handler which receives key events when it has focus.
func (m *Manager) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return m.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		m.Lock()
		defer m.Unlock()

		inputHandler := m.visibleRoot.InputHandler()
		if inputHandler != nil {
			inputHandler(event, setFocus)
		}
	})
}
