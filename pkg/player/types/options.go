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
	cfg.PathToMPV = string(opt)
}

type OptionLowLatency bool

func (opt OptionLowLatency) Apply(cfg *Config) {
	cfg.LowLatency = bool(opt)
}

type OptionCacheDuration time.Duration

func (opt OptionCacheDuration) Apply(cfg *Config) {
	if opt < 0 {
		cfg.CacheLength = nil
		return
	}
	cfg.CacheLength = ptr(time.Duration(opt))
}

type OptionCacheMaxSize uint64

func (opt OptionCacheMaxSize) Apply(cfg *Config) {
	cfg.CacheMaxSize = uint64(opt)
}
