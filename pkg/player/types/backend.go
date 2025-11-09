package types

type Backend string

const (
	BackendUndefined       = ""
	BackendLibVLC          = "libvlc"
	BackendMPV             = "mpv"
	BackendGStreamerEbiten = "gstreamer_ebiten"
	BackendGStreamerFyne   = "gstreamer_fyne"
	BackendLibAVEbiten     = "libav_ebiten"
	BackendLibAVFyne       = "libav_fyne"
)
