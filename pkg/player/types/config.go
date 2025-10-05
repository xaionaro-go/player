package types

import "time"

type Preset string

const (
	PresetLowLatency = Preset("low_latency")
)

type Config struct {
	PathToMPV    *string        `yaml:"path_to_mpv"`
	Preset       *Preset        `yaml:"preset"`
	AudioBuffer  *time.Duration `yaml:"audio_buffer"`
	CacheLength  *time.Duration `yaml:"cache_length"`
	CacheMaxSize *uint64        `yaml:"cache_max_size"`
	HideWindow   bool           `yaml:"hide_window"`
}
