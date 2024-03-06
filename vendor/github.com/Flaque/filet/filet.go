package filet

import (
	"bytes"
	"github.com/spf13/afero"
	"os"
	"path/filepath"
)

// TestReporter can be used to report test failures. It is satisfied by the standard library's *testing.T.
type TestReporter interface {
	Error(args ...interface{})
}

// Files keeps track of files that we've used so we can clean up.
var Files []string
var appFs = afero.NewOsFs()

/*
TmpDir Creates a tmp directory for us to use.
*/
func TmpDir(t TestReporter, dir string) string {
	name, err := afero.TempDir(appFs, dir, "dir")
	if err != nil {
		t.Error("Failed to create the tmpDir: "+name, err)
	}

	Files = append(Files, name)
	return name
}

/*
TmpFile Creates a tmp file for us to use when testing
*/
func TmpFile(t TestReporter, dir string, content string) afero.File {
	file, err := afero.TempFile(appFs, dir, "file")

	if err != nil {
		t.Error("Failed to create the tmpFile: "+file.Name(), err)
	}

	file.WriteString(content)
	Files = append(Files, file.Name())

	return file
}

/*
File Creates a specified file for us to use when testing
*/
func File(t TestReporter, path string, content string) afero.File {
	file, err := appFs.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		t.Error("Failed to create the file: "+path, err)
		return nil
	}

	file.WriteString(content)
	Files = append(Files, file.Name())

	return file
}

/*
FileSays returns true if the file at the path contains the expected byte array
content.
*/
func FileSays(t TestReporter, path string, expected []byte) bool {
	content, err := afero.ReadFile(appFs, path)
	if err != nil {
		t.Error("Failed to read file: "+path, err)
	}

	return bytes.Equal(content, expected)
}

/*
CleanUp removes all files in our test registry and calls `t.Error` if something goes
wrong.
*/
func CleanUp(t TestReporter) {
	for _, path := range Files {
		if err := appFs.RemoveAll(path); err != nil {
			t.Error(appFs.Name(), err)
		}
	}

	Files = make([]string, 0)
}

/*
Exists returns true if the file exists. Calls t.Error if something goes wrong while
checking.
*/
func Exists(t TestReporter, path string) bool {
	exists, err := afero.Exists(appFs, path)
	if err != nil {
		t.Error("Something went wrong when checking if "+path+"exists!", err)
	}
	return exists
}

/*
DirContains returns true if the dir contains the path. Calls t.Error if
something goes wrong while checking.
*/
func DirContains(t TestReporter, dir string, path string) bool {
	fullPath := filepath.Join(dir, path)
	return Exists(t, fullPath)
}