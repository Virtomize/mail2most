//go:build mage
// +build mage

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/magefile/mage/sh"
	"github.com/mholt/archiver"
)

var (
	binPath        = "bin"
	binName        = "mail2most"
	confFile       = "conf/mail2most.conf"
	dockerPath     = "docker"
	dockerFilePath = "Dockerfile"
	registry       = "virtomize/mail2most"
)

// Build - mage build
func Build() error {
	err := Clean()
	if err != nil {
		return err
	}

	err = Test()
	if err != nil {
		return err
	}

	tag, _ := exec.Command("bash", "-c", "git tag --sort=-version:refname | head -n 1").Output()

	return sh.RunV("go", "build", "-a", "-tags", "netgo", "-o", binPath+"/"+binName, "-ldflags", "-w -extldflags \"-static\" -X 'main.Version="+string(tag)+"'")
}

// CreateRelease - mage build
func CreateRelease() error {
	err := Clean()
	if err != nil {
		return err
	}

	err = Test()
	if err != nil {
		return err
	}
	osarch := make(map[string][]string)
	osarch["linux"] = []string{"386", "amd64", "arm", "arm64"}
	osarch["windows"] = []string{"386", "amd64"}

	for goos, archs := range osarch {
		for _, arch := range archs {
			path := fmt.Sprintf("%s/%s/", binPath, goos+"-"+arch)

			err := os.Setenv("GOOS", goos)
			if err != nil {
				return err
			}

			err = os.Setenv("GOARCH", arch)
			if err != nil {
				return err
			}

			tag, _ := exec.Command("bash", "-c", "git tag --sort=-version:refname | head -n 1").Output()

			err = sh.RunV("go", "build", "-a", "-tags", "netgo", "-o", path+binName, "-ldflags", "-w -extldflags \"-static\" -X 'main.Version="+string(tag)+"'")
			if err != nil {
				return err
			}

			err = os.Mkdir(path+"conf", 0755)
			if err != nil {
				return err
			}

			err = copyFile(confFile, path+confFile, 2000)
			if err != nil {
				return err
			}

			err = archiver.Archive([]string{path + binName, path + confFile}, binPath+"/"+goos+"-"+arch+".tar.gz")
			if err != nil {
				return err
			}

		}
	}
	return nil
}

// Test - running tests and code coverage
func Test() error {
	return sh.RunV("go", "test", "-v", "-cover", "./...", "-coverprofile=coverage.out")
}

// Run - mage run
func Run() error {
	return sh.RunV("go", "run", "main.go")
}

// Coverage - checking code coverage
func Coverage() error {
	if _, err := os.Stat("./coverage.out"); err != nil {
		return fmt.Errorf("run mage test befor checking the code coverage")
	}
	return sh.RunV("go", "tool", "cover", "-html=coverage.out")
}

// Clean cleans up the client generation and binarys
func Clean() error {
	fmt.Println("cleaning up")
	if _, err := os.Stat("coverage.out"); err == nil {
		err = os.Remove("coverage.out")
		if err != nil {
			return err
		}
	}
	err := cleanDocker()
	if err != nil {
		return err
	}
	return os.RemoveAll("bin/")
}

// Docker creates docker container
func Docker() error {
	if err := cleanDocker(); err != nil {
		return err
	}

	if err := os.MkdirAll(dockerPath+"/conf", 0755); err != nil {
		return err
	}

	if err := sh.RunWith(map[string]string{"CGO_ENABLED": "0"}, "go", "build", "-a", "-installsuffix", "cgo", "-o", "docker/mail2most", "main.go"); err != nil {
		return err
	}

	if err := copyFile(dockerFilePath, dockerPath+"/Dockerfile", 1000); err != nil {
		return err
	}

	copyFile(confFile, dockerPath+"/"+confFile, 1000)

	cmd := exec.Command("git", "describe", "--tags")
	b, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	tag := strings.TrimSuffix(string(b), "\n")

	if err := sh.RunV("docker", "build", "-t", registry+":"+tag, "docker"); err != nil {
		return err
	}

	if err := sh.RunV("docker", "push", registry+":"+tag); err != nil {
		return err
	}

	if err := sh.RunV("docker", "build", "-t", registry+":latest", "docker"); err != nil {
		return err
	}

	return sh.RunV("docker", "push", registry+":latest")
}

func cleanDocker() error {
	if _, err := os.Stat(dockerPath); err == nil {
		err = os.RemoveAll(dockerPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string, BUFFERSIZE int64) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	_, err = os.Stat(dst)
	if err == nil {
		return fmt.Errorf("File %s already exists", dst)
	}

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	if err != nil {
		panic(err)
	}

	buf := make([]byte, BUFFERSIZE)
	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		if _, err := destination.Write(buf[:n]); err != nil {
			return err
		}
	}
	return err
}
