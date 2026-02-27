package wizard

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockDetector implements Detector for testing.
type mockDetector struct {
	binaries map[string]bool
	files    map[string]bool
	dirs     map[string]bool
}

func (m *mockDetector) LookPath(name string) (string, error) {
	if m.binaries[name] {
		return "/usr/bin/" + name, nil
	}
	return "", &os.PathError{Op: "lookpath", Path: name, Err: os.ErrNotExist}
}

type fakeFileInfo struct {
	name  string
	isDir bool
}

func (f fakeFileInfo) Name() string        { return f.name }
func (f fakeFileInfo) Size() int64         { return 0 }
func (f fakeFileInfo) Mode() os.FileMode   { return 0644 }
func (f fakeFileInfo) ModTime() time.Time  { return time.Time{} }
func (f fakeFileInfo) IsDir() bool         { return f.isDir }
func (f fakeFileInfo) Sys() interface{}    { return nil }

func (m *mockDetector) Stat(path string) (os.FileInfo, error) {
	if m.dirs[path] {
		return fakeFileInfo{name: path, isDir: true}, nil
	}
	if m.files[path] {
		return fakeFileInfo{name: path, isDir: false}, nil
	}
	return nil, os.ErrNotExist
}

func (m *mockDetector) Glob(pattern string) ([]string, error) {
	return nil, nil
}

func TestDetectTailscale(t *testing.T) {
	d := &mockDetector{binaries: map[string]bool{"tailscale": true}}
	result := Detect(d)
	assert.True(t, result.TailscaleAvailable)
}

func TestDetectNoTailscale(t *testing.T) {
	d := &mockDetector{binaries: map[string]bool{}}
	result := Detect(d)
	assert.False(t, result.TailscaleAvailable)
}

func TestDetectAnsibleInventory(t *testing.T) {
	d := &mockDetector{
		binaries: map[string]bool{},
		files:    map[string]bool{"inventory/hosts.yml": true},
	}
	result := Detect(d)
	assert.Equal(t, "inventory/hosts.yml", result.AnsibleInventory)
}

func TestDetectComposeFiles(t *testing.T) {
	d := &mockDetector{
		binaries: map[string]bool{},
		files:    map[string]bool{"docker-compose.yml": true, "compose.yml": true},
	}
	result := Detect(d)
	assert.Contains(t, result.ComposeFiles, "docker-compose.yml")
	assert.Contains(t, result.ComposeFiles, "compose.yml")
}

func TestDetectNothing(t *testing.T) {
	d := &mockDetector{
		binaries: map[string]bool{},
		files:    map[string]bool{},
	}
	result := Detect(d)
	assert.False(t, result.TailscaleAvailable)
	assert.Empty(t, result.AnsibleInventory)
	assert.Empty(t, result.ComposeFiles)
}
