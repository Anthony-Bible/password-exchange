// Package buildinfo exposes build-time metadata injected via -ldflags.
package buildinfo

// Version is the application version, set at build time via:
//
//	-ldflags "-X github.com/Anthony-Bible/password-exchange/app/internal/shared/buildinfo.Version=<version>"
//
// It defaults to "dev" for local builds without ldflags.
var Version = "dev"
