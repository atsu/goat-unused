package fio

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"syscall"
	"time"
)

const (
	linuxCacheDropPath = "/proc/sys/vm/drop_caches"
)

func RunDefaultModeOperations(mode, directory string, sleep time.Duration) error {
	switch mode {
	case "summary":
		runAllOperations(directory)
	case "collect":
		runAllOperations(directory)
	case "debug":
		runAllOperations(directory)
	case "manifest":
		runAllOperations(directory)
	default:
		return fmt.Errorf("invalid mode %s", mode)
	}

	time.Sleep(sleep)
	return nil
}

func runAllOperations(directoryPath string) {
	/*
		Some operation unaccounted for, due to cache, or inability to produce.
		mount - ignored for now
		readlink - requires link to pre-exist, or cache break
		readdir - seems like readdirplus is always used instead
		remove - ?
		rename - ?
	*/

	directory := createTempDir(directoryPath)
	defer removeDirectory(directory)
	log.Printf("Target Directory: %v\n", directory)

	CreateFileWithData(path.Join(directory, "bar1.foo"), []byte("Test Data"), true)
	renameFile(path.Join(directory, "bar1.foo"), path.Join(directory, "bar.foo"))
	statFile(path.Join(directory, "bar.foo"))
	lStatFile(path.Join(directory, "bar.foo"))
	chmodFile(path.Join(directory, "bar.foo"))
	linkFile(path.Join(directory, "bar.foo"), path.Join(directory, "foo.link"))
	ReadFile(path.Join(directory, "foo.link"), true)
	unlinkFile(path.Join(directory, "foo.link"))
	symLinkFile(path.Join(directory, "bar.foo"), path.Join(directory, "foo.link"))
	ReadLink(path.Join(directory, "foo.link"), true)
	ReadDirectory(directory, true)
}

func createTempDir(directoryPath string) string {
	pid := os.Getpid()
	now := strconv.FormatInt(time.Now().Unix(), 10)
	tempDirName, err := ioutil.TempDir(directoryPath, fmt.Sprint(now, "-", pid))
	if err != nil {
		log.Fatal(err)
	}
	return tempDirName
}

func statFile(fullFilePath string) {
	dropLinuxCache()
	log.Printf("Stats: %v\n", fullFilePath)
	if finfo, err := os.Stat(fullFilePath); err != nil {
		log.Fatal(err)
	} else {
		log.Printf(" -Dir: %v, FileName: %v, Size: %v\n", finfo.IsDir(), finfo.Name(), finfo.Size())
	}
}

func lStatFile(fullFilePath string) {
	dropLinuxCache()
	log.Printf("Lstats: %v\n", fullFilePath)
	if finfo, err := os.Lstat(fullFilePath); err != nil {
		log.Fatal(err)
	} else {
		log.Printf(" -Dir: %v, FileName: %v, Size: %v\n", finfo.IsDir(), finfo.Name(), finfo.Size())
	}
}

func linkFile(fileSource, fileTarget string) {
	log.Printf("Linking: %v -> %v\n", fileSource, fileTarget)
	if err := os.Link(fileSource, fileTarget); err != nil {
		log.Fatal(err)
	}
}

func symLinkFile(fileSource, fileTarget string) {
	log.Printf("Symlinking: %v -> %v\n", fileSource, fileTarget)
	if err := os.Symlink(fileSource, fileTarget); err != nil {
		log.Fatal(err)
	}
}

func removeFile(fullFilePath string) {
	log.Printf("Removing file: %v\n", fullFilePath)
	if err := os.Remove(fullFilePath); err != nil {
		log.Fatal(err)
	}
}

func dropLinuxCache() {
	if os.Geteuid() == 0 {
		if _, err := os.Stat(linuxCacheDropPath); os.IsNotExist(err) {
			log.Fatal(err)
		}

		f, err := os.OpenFile(linuxCacheDropPath, os.O_WRONLY, os.ModeExclusive)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		if _, err := f.Write([]byte("3")); err != nil {
			log.Fatal(err)
		}
	}
}

func unlinkFile(fullFilePath string) {
	if err := syscall.Unlink(fullFilePath); err != nil {
		log.Fatal(err)
	}
}

func removeDirectory(fullDirectoryPath string) {
	log.Printf("Removing Directory: %v\n", fullDirectoryPath)
	if err := os.RemoveAll(fullDirectoryPath); err != nil {
		log.Fatal(err)
	}
}

func renameFile(fileSource, fileTarget string) {
	log.Printf("Renaming file: %v -> %v\n", fileSource, fileTarget)
	if err := os.Rename(fileSource, fileTarget); err != nil {
		log.Fatal(err)
	}
}

func chmodFile(fullFilePath string) {
	dropLinuxCache()
	if err := os.Chmod(fullFilePath, 0777); err != nil {
		log.Fatal(err)
	}
}
