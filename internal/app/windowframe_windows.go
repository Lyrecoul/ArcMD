package app

import (
	"unsafe"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"golang.org/x/sys/windows"
)

var setDWMAttribute = windows.NewLazySystemDLL("dwmapi.dll").NewProc("DwmSetWindowAttribute")

func newAppWindow(opts mygo.WindowOptions, view func(*ui.Context)) *mygo.Window {
	opts.Content = ui.View(view)
	win := mygo.NewWindow(opts)
	hwnd := win.NativeHandle()
	if hwnd == 0 || setDWMAttribute.Find() != nil {
		return win
	}
	set := func(attribute uintptr, value uint32) bool {
		result, _, _ := setDWMAttribute.Call(hwnd, attribute, uintptr(unsafe.Pointer(&value)), unsafe.Sizeof(value))
		return int32(result) >= 0
	}
	const dwmCornerPreference, dwmBorderColor = 33, 34
	const dwmRound, dwmDefaultColor = 2, 0xffffffff
	if set(dwmCornerPreference, dwmRound) {
		update := func() {
			if !win.IsFocused() {
				set(dwmBorderColor, dwmDefaultColor)
				return
			}
			c := windowFrameInk(mygo.Theme.IsDark())
			set(dwmBorderColor, uint32(c.R)|uint32(c.G)<<8|uint32(c.B)<<16)
		}
		win.OnFocus(update)
		win.OnBlur(update)
		win.OnClosed(mygo.Theme.OnUpdated(update))
		update()
	}
	return win
}

func hasAppWindowControls() bool { return false }
