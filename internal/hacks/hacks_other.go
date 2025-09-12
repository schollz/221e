//go:build !windows
// +build !windows

package hacks

func StoreWinClipboard() error {
	return nil
}
