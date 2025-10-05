package types

import "time"

type Option interface {
	Apply(cfg *Config)
}

type Options []Option

func (options Options) Config() Config {
	cfg := Config{}
	options.Apply(&cfg)
	return cfg
}

func (options Options) Apply(cfg *Config) {
	for _, option := range options {
		option.Apply(cfg)
	}
}

type OptionPathToMPV string

func (opt OptionPathToMPV) Apply(cfg *Config) {
	cfg.PathToMPV = ptr(string(opt))
}

type OptionNoPathToMPV struct{}

func (opt OptionNoPathToMPV) Apply(cfg *Config) {
	cfg.PathToMPV = nil
}

type OptionPreset Preset

func (opt OptionPreset) Apply(cfg *Config) {
	cfg.Preset = ptr(Preset(opt))
}

type OptionNoPreset struct{}

func (opt OptionNoPreset) Apply(cfg *Config) {
	cfg.Preset = nil
}

type OptionAudioBuffer time.Duration

func (opt OptionAudioBuffer) Apply(cfg *Config) {
	if opt < 0 {
		panic("audio buffer must be non-negative")
	}
	cfg.AudioBuffer = ptr(time.Duration(opt))
}

type OptionNoAudioBuffer struct{}

func (opt OptionNoAudioBuffer) Apply(cfg *Config) {
	cfg.AudioBuffer = nil
}

type OptionCacheDuration time.Duration

func (opt OptionCacheDuration) Apply(cfg *Config) {
	if opt < 0 {
		panic("cache duration must be non-negative")
	}
	cfg.CacheLength = ptr(time.Duration(opt))
}

type OptionNoCacheDuration struct{}

func (opt OptionNoCacheDuration) Apply(cfg *Config) {
	cfg.CacheLength = nil
}

type OptionCacheMaxSize uint64

func (opt OptionCacheMaxSize) Apply(cfg *Config) {
	cfg.CacheMaxSize = ptr(uint64(opt))
}

type OptionNoCacheMaxSize struct{}

func (opt OptionNoCacheMaxSize) Apply(cfg *Config) {
	cfg.CacheMaxSize = nil
}

type OptionHideWindow bool

func (opt OptionHideWindow) Apply(cfg *Config) {
	cfg.HideWindow = bool(opt)
}
