//go:build !linux

package assets

// RunSubStoreService is a no-op on non-Linux platforms.
func RunSubStoreService() {}
