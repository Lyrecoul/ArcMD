package app

import (
	"os"
	"strings"
	"sync"

	"github.com/ebitengine/purego"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

type linuxFrameSupport struct {
	state func(uintptr) uint32
	kwin  bool
}

var linuxFrame = sync.OnceValue(func() *linuxFrameSupport {
	// GTK reports the compositor's default, not whether it accepts CSD.
	gdk, err := purego.Dlopen("libgdk-3.so.0", purego.RTLD_LAZY|purego.RTLD_LOCAL)
	if err != nil {
		return nil
	}
	var displayDefault, screenDefault func() uintptr
	var rgbaVisual func(uintptr) uintptr
	var composited func(uintptr) int32
	var state func(uintptr) uint32
	purego.RegisterLibFunc(&displayDefault, gdk, "gdk_display_get_default")
	purego.RegisterLibFunc(&screenDefault, gdk, "gdk_screen_get_default")
	purego.RegisterLibFunc(&rgbaVisual, gdk, "gdk_screen_get_rgba_visual")
	purego.RegisterLibFunc(&composited, gdk, "gdk_screen_is_composited")
	purego.RegisterLibFunc(&state, gdk, "gdk_window_get_state")
	display, screen := displayDefault(), screenDefault()
	if display == 0 || screen == 0 || rgbaVisual(screen) == 0 || composited(screen) == 0 {
		return nil
	}
	kwin := false
	if typ, err := purego.Dlsym(gdk, "gdk_wayland_display_get_type"); err == nil {
		var waylandType func() uintptr
		var isType func(uintptr, uintptr) int32
		purego.RegisterFunc(&waylandType, typ)
		purego.RegisterLibFunc(&isType, gdk, "g_type_check_instance_is_a")
		if isType(display, waylandType()) != 0 {
			sym, err := purego.Dlsym(gdk, "gdk_wayland_display_prefers_ssd")
			if err != nil {
				return nil
			}
			var prefersSSD func(uintptr) int32
			purego.RegisterFunc(&prefersSSD, sym)
			if prefersSSD(display) != 0 {
				// KWin accepts MyGo's pre-map client-decoration request for
				// frameless windows. Other SSD-first compositors can refuse it.
				for desktop := range strings.SplitSeq(os.Getenv("XDG_CURRENT_DESKTOP"), ":") {
					kwin = kwin || strings.EqualFold(desktop, "KDE")
				}
				if !kwin {
					return nil
				}
			}
		}
	}
	gtk, err := purego.Dlopen("libgtk-3.so.0", purego.RTLD_LAZY|purego.RTLD_LOCAL)
	if err != nil {
		return nil
	}
	var widgetWindow func(uintptr) uintptr
	purego.RegisterLibFunc(&widgetWindow, gtk, "gtk_widget_get_window")
	return &linuxFrameSupport{kwin: kwin, state: func(widget uintptr) uint32 {
		if widget != 0 {
			if window := widgetWindow(widget); window != 0 {
				return state(window)
			}
		}
		return 0
	}}
})

func newAppWindow(opts mygo.WindowOptions, view func(*ui.Context)) *mygo.Window {
	frame := linuxFrame()
	if frame == nil {
		opts.Content = ui.View(view)
		return mygo.NewWindow(opts)
	}
	// Use MyGo's native resize hit targets in either mode. Frameless is
	// needed on KWin to request CSD before the window is mapped.
	opts.Transparent = true
	opts.Frameless = frame.kwin
	// Frameless bypasses MyGo's hidden-title-bar menu policy. Keep the
	// native menu available on Alt/F10 without a permanent row above us.
	opts.AutoHideMenuBar = opts.AutoHideMenuBar || frame.kwin
	var widget uintptr
	opts.Content = ui.View(func(c *ui.Context) {
		const attached = 1<<2 | 1<<4 | 1<<8 | 1<<9 | 1<<11 | 1<<13 | 1<<15 // maximized, fullscreen, tiled edges
		var radius float32
		status := frame.state(widget)
		if status&attached == 0 {
			radius = 12
		}
		c.Root().Radius(radius).Clip()
		view(c)
		ui.Box(c.Key("window-frame-outline")).Absolute().Top(0).Right(0).Bottom(0).Left(0).
			Radius(radius).Border(1, separator(c.Theme())).PassThrough()
	})
	win := mygo.NewWindow(opts)
	widget = win.NativeHandle()
	return win
}

func hasAppWindowControls() bool {
	frame := linuxFrame()
	return frame != nil && frame.kwin
}
