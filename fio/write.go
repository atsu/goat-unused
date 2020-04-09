package fio

import (
	"io"
	"log"
	"os"
)

func MkDir(dir string, v bool) error {
	if v {
		log.Printf("Making Directory %v", dir)
	}

	if err := os.Mkdir(dir, 0755); err != nil {
		return err
	}

	return nil
}

func CreateJunkFile(file string, size uint64, v bool) (uint64, error) {
	if v {
		log.Printf("Creating file: %v\n", file)
	}
	f, err := os.OpenFile(file, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return 0, err
	}
	if f != nil {
		defer f.Close()
	}

	bc := uint64(0)

	for bc < size {
		b, e := GenerateRandomBytes(uint64(os.Getpagesize()))
		if e != nil {
			return 0, err
		}
		wc, err := f.Write(b)
		if err != nil {
			return 0, err
		}
		bc += uint64(wc)
	}

	if v {
		log.Printf(" -Bytes Writtern: %v\n", bc)
	}
	return bc, nil
}

func CreateFileWithData(file string, data []byte, v bool) (int, error) {
	if v {
		log.Printf("Creating file with Data: %v\n", file)
	}
	f, err := os.Create(file)
	if err != nil {
		log.Fatal(err)
	}
	if f != nil {
		defer f.Close()
	}

	n, err := f.Write(data)
	if err != nil {
		return 0, err
	}

	if v {
		log.Printf(" -Bytes Written: %v\n", n)
	}
	return n, nil
}

func CreateFileWithStream(file string, reader io.Reader, v bool) (int64, error) {
	if v {
		log.Printf("Creating file: %v\n", file)
	}
	f, err := os.OpenFile(file, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return 0, err
	}
	if f != nil {
		defer f.Close()
	}

	n, err := io.Copy(f, reader)
	if err != nil {
		return 0, err
	}

	if v {
		log.Printf(" -Bytes Writtern: %v\n", n)
	}
	return n, nil
}

func WriteToFile(file string, data []byte, v bool) (int, error) {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		log.Fatal(err)
	}
	if f != nil {
		defer f.Close()
	}

	n, err := f.Write(data)
	if err != nil {
		return 0, err
	}
	if v {
		log.Printf(" -Bytes Written: %v\n", n)
	}
	return n, nil
}
