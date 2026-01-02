// Package adapters provides the handler registry and built-in handlers.
package adapters

import (
	"github.com/ensurascript/ensura/pkg/adapters/aes"
	"github.com/ensurascript/ensura/pkg/adapters/cron"
	"github.com/ensurascript/ensura/pkg/adapters/fs"
	"github.com/ensurascript/ensura/pkg/adapters/http"
	"github.com/ensurascript/ensura/pkg/adapters/posix"
	"github.com/ensurascript/ensura/pkg/runtime"
)

// NewDefaultRegistry creates a registry with all built-in handlers.
func NewDefaultRegistry() *runtime.HandlerRegistry {
	registry := runtime.NewHandlerRegistry()

	// Register filesystem handler
	registry.Register(fs.New())

	// Register POSIX permissions handler
	registry.Register(posix.New())

	// Register AES encryption handler
	registry.Register(aes.New())

	// Register HTTP handler
	registry.Register(http.New())

	// Register cron handler
	registry.Register(cron.New())

	return registry
}
