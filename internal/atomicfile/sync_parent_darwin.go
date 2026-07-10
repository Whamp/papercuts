//go:build darwin

package atomicfile

// syncParent cannot portably flush a directory entry on Darwin. Go's File.Sync
// uses F_FULLFSYNC there, which directory descriptors may reject. The complete
// temporary file is synced before the atomic publication or replacement.
func syncParent(string) error {
	return nil
}
