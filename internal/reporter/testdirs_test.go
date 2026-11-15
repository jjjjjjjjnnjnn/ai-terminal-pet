package reporter

import (
	"os"
	"path/filepath"
	"testing"

	"ai-terminal-pet/internal/home"
)

// One test helper, so no test has to know which environment variable decides
// where this project's data directory is on the machine the suite runs on.
//
// It exists because of a mistake worth recording: the tests used to set
// LOCALAPPDATA and nothing else, which was correct while the path was spelled by
// hand and became wrong the moment internal/home learned three platforms. Left
// unfixed, the linux run wrote snapshots into the tester's real
// ~/.local/share/ai-terminal-pet and the next assertion found them there.

// dataDirectory points every variable a platform might use at one temp tree and
// returns it. The directory the code then resolves is inside that tree.
func dataDirectory(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv("LOCALAPPDATA", filepath.Join(root, "local"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(root, "xdg"))
	t.Setenv("HOME", filepath.Join(root, "profile"))
	t.Setenv("USERPROFILE", filepath.Join(root, "profile"))
	if directory := home.Directory(); directory == "" {
		t.Fatal("the platform resolved no data directory from the test environment")
	}
	return root
}

// dataDirectoryMissing clears every variable that could give the overlay a
// directory, so the code under test has to answer without one.
func dataDirectoryMissing(t *testing.T) {
	t.Helper()
	for _, name := range []string{"LOCALAPPDATA", "XDG_DATA_HOME", "HOME", "USERPROFILE"} {
		t.Setenv(name, "")
	}
}

// osReadDirNames is the tiny helper kept next to the others so a test can assert
// "the directory is empty" without importing two packages.
func osReadDirNames(t *testing.T, directory string) []string {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}
