package types

import "time"

type Config struct {
	PathToMPV    string         `yaml:"path_to_mpv"`
	LowLatency   bool           `yaml:"low_latency"`
	CacheLength  *time.Duration `yaml:"cache_length"`
	CacheMaxSize uint64         `yaml:"cache_max_size"`
	HideWindow   bool           `yaml:"hide_window"`
}
