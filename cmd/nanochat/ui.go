package main

import (
	"fmt"

	"github.com/gabriel-comeau/tbuikit"
	"github.com/nsf/termbox-go"
)

///////////////////
// UI dimensions //
///////////////////

const (
	titleBarHeight  = 3
	chatInputHeight = 3
)

func calcTopTitleBar() (x1, x2, y1, y2 int) {
	x1 = 0
	x2 = tbuikit.GetTermboxWidth() - 1
	y1 = 0
	y2 = titleBarHeight - 1
	return
}

func calcChatMsgs() (x1, x2, y1, y2 int) {
	w, h := termbox.Size()
	x1 = 0
	x2 = w - 1
	y1 = titleBarHeight
	y2 = h - chatInputHeight - 1
	return
}

func calcChatInput() (x1, x2, y1, y2 int) {
	w, h := termbox.Size()
	x1 = 0
	x2 = w - 1
	y1 = h - chatInputHeight
	y2 = h - 1
	return
}

/////////////
// UI type //
/////////////

type chatUI struct {
	msgBuffer   *tbuikit.StringBuffer
	inputBuffer *tbuikit.TextInputBuffer
	ui          *tbuikit.UI
	screen      *tbuikit.Screen
	inputMsgs   chan (string)
	chatMsgs    chan (string)
}

// newChatUI creates and initializes a new chat UI
func newChatUI() *chatUI {
	cui := new(chatUI)

	// Handler for quit event
	quitHandler := func(interface{}, interface{}) {
		cui.ui.Shutdown()
	}

	// Handler called when user presses the enter key
	chatInputEnter := func(uiElement, event interface{}) {
		widget, ok := uiElement.(*tbuikit.TextInputWidget)
		if ok {
			go func() {
				msg := widget.GetBuffer().ReturnAndClear()
				cui.inputMsgs <- msg
				cui.chatMsgs <- fmt.Sprintf("<Me:> %v\n", msg)
			}()
		}
	}

	// Initialize buffers and channels
	cui.msgBuffer = new(tbuikit.StringBuffer)
	cui.msgBuffer.Prepare(64)
	cui.inputBuffer = new(tbuikit.TextInputBuffer)
	cui.inputMsgs = make(chan string)
	cui.chatMsgs = make(chan string)

	// Initialize UI objects
	cui.ui = new(tbuikit.UI)
	cui.screen = new(tbuikit.Screen)

	// Initialize screen widgets:

	// Title Bar
	titleBar := tbuikit.CreateLabelWidget(
		"Nanochat - press ESC/Ctrl-C to quit", true, tbuikit.CENTER,
		termbox.ColorGreen, termbox.ColorDefault, termbox.ColorWhite,
		calcTopTitleBar)

	// Chat message box
	chatMsgs := tbuikit.CreateStringDisplayWidget(termbox.ColorWhite, termbox.ColorWhite,
		termbox.ColorDefault, calcChatMsgs, cui.msgBuffer)

	// New message input box
	chatInput := tbuikit.CreateTextInputWidget(true, termbox.ColorWhite,
		termbox.ColorDefault, termbox.ColorDefault, calcChatInput,
		cui.inputBuffer, true, true)
	chatInput.UseDefaultKeys(true)
	chatInput.AddSpecialKeyCallback(termbox.KeyEnter, chatInputEnter)

	// Add widgets to screen
	cui.screen.AddWidget(titleBar)
	cui.screen.AddWidget(chatMsgs)
	cui.screen.AddWidget(chatInput)

	// Key bindings
	cui.screen.AddSpecialKeyCallback(termbox.KeyEsc, quitHandler)
	cui.screen.AddSpecialKeyCallback(termbox.KeyCtrlC, quitHandler)

	// Add chat screen and activate it
	cui.ui.AddScreen(cui.screen)
	cui.screen.Activate()

	return cui
}

// run runs the UI event loop in a separate goroutine (non-blocking)
func (cui *chatUI) run(quit chan bool) {
	go func() {
		for {
			msg := <-cui.chatMsgs
			cui.msgBuffer.Add(msg + "\n")
		}
	}()

	go cui.ui.Start(quit)
}
