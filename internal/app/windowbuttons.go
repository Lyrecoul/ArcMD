package app

import (
	"github.com/Ciao1019/ArcMD/internal/i18n"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

const windowButtonsWidth = 3*30 + 2*4

func winNeedsCompactToolbar(c *ui.Context, docked float32, win *mygo.Window) bool {
	w, _ := c.Size()
	return w-docked < 400 && windowButtonsVisible(win)
}

func windowButtonsVisible(win *mygo.Window) bool {
	return win != nil && hasAppWindowControls() && !win.IsFullScreen()
}

// windowButtons adds client-owned controls to an existing toolbar, rather
// than creating a title row above the application.
func windowButtons(c *ui.Context, win *mygo.Window) {
	if !windowButtonsVisible(win) {
		return
	}
	ui.Row(c.Key("window-controls")).Shrink(0).PaddingX(4).AlignItems(ui.Center).Children(func() {
		caption := func(key string, icon *ui.SVG, label string, action func()) {
			label = i18n.T(label)
			if iconButton(c, key, icon, label, label, false).Shrink(0).Clicked() {
				action()
			}
		}
		caption("minimize", iconMinus, "最小化窗口", win.Minimize)
		if win.IsMaximized() {
			caption("maximize", iconWindowRestore, "还原窗口", win.ToggleMaximize)
		} else {
			caption("maximize", iconWindowFrame, "最大化窗口", win.ToggleMaximize)
		}
		caption("close", iconX, "关闭窗口", win.Close)
	})
}
