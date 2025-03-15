package main

import (
	_ "embed"
	"syscall"
	
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/pango"
)

//go:embed glade/17_MainWindow.glade
var WindowGlade string

//go:embed resources/sticky-note.ico
var icon []byte

type StickyStr struct {
	X				int
	Y				int
	Width			int
	Height			int
	Note			string
	FontFamily		string
	FontSize		int
	FontStyle		pango.Style
	FontWeight		int
	FontStretch		pango.Stretch
	FgColor			[4]int
	BgColor			[4]int
	Align			float64
	stickyWindow	*gtk.Window
}

const (
	WS_EX_TOOLWINDOW = 0x00000080
	GWL_EXSTYLE      = -20
)

const (
	STICKY_NEW = iota
	STICKY_UPDATED
	STICKY_NOT_UPDATED
	EDIT_NEW
	EDIT_UPDATED
)

var (
	User32				= syscall.NewLazyDLL("user32.dll")
	GetWindowLongPtr	= User32.NewProc("GetWindowLongPtrW")
	SetWindowLongPtr	= User32.NewProc("SetWindowLongPtrW")
)

var MyApplication	*gtk.Application
var StickyMap		map[string]StickyStr
var ConfFile		string
