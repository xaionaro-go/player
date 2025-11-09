package main

import (
	"context"

	"github.com/facebookincubator/go-belt/tool/logger"
)

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func try[T any](v T, err error) T {
	if err != nil {
		logger.Errorf(context.Background(), "%s", err)
	}
	return v
}
