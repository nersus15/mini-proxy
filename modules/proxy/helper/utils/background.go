package utils

import (
	"runtime/debug"

	"github.com/webcore-go/webcore/infra/logger"
)

func RunBackground(name string, fn func()) {
	go func() {
		logger.Info("Start Job: " + name + " di background")
		defer RecoverBackground(name)
		fn()
	}()
}

func RecoverBackground(name string) {
	if r := recover(); r != nil {
		logger.Error("Panic terjadi di background "+name, "err", r, "stack", string(debug.Stack()))
	}
}
