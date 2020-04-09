package fio

import (
	"io/ioutil"
	"log"
	"os"
)

func ReadFile(file string, v bool) ([]byte, error) {
	if v {
		log.Printf("Reading File: %v\n", file)
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	if v {
		log.Printf(" -Read %d bytes of data", len(data))
	}

	return data, nil

}

func ReadDirectory(dir string, v bool) ([]os.FileInfo, error) {
	if v {
		log.Printf("Reading Dir: %v\n", dir)
	}

	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	if v {
		for _, info := range infos {
			log.Printf(" -Dir: %v, FileName: %v, Size: %v\n", info.IsDir(), info.Name(), info.Size())
		}
	}

	return infos, nil
}

func ReadLink(link string, v bool) (string, error) {
	if v {
		log.Printf("Reading Link: %v\n", link)
	}

	link, err := os.Readlink(link)
	if err != nil {
		return "", err
	}
	if v {
		log.Printf(" -Linked: %v\n", link)
	}
	return link, nil
}
