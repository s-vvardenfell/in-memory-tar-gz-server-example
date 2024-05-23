package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"time"
)

func Tar(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	// creating tar writer from new buffer
	tw := tar.NewWriter(&buf)
	defer tw.Close()

	// manually create tar header
	hdr := &tar.Header{
		Name:     "file", // use your value if need
		Size:     int64(len(data)),
		Mode:     509,
		ModTime:  time.Now(),
		Typeflag: tar.TypeReg, // regular file
	}

	err := tw.WriteHeader(hdr)
	if err != nil {
		return nil, err
	}

	num, err := tw.Write(data)
	if err != nil {
		return nil, err
	}

	// check if all data written
	if num == 0 || num != len(data) {
		return nil, errors.New("tar wrote zero or wrong num of bytes")
	}

	return buf.Bytes(), nil
}

// you may use io.Reader as an argument
func Untar(data []byte) ([]byte, error) {
	tr := tar.NewReader(bytes.NewReader(data))
	var outBuf bytes.Buffer // buffer for data

	for {
		_, err := tr.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		num, err := io.Copy(&outBuf, tr)
		if err != nil {
			return nil, err
		}

		if num == 0 {
			return nil, errors.New("untar copy zero bytes")
		}
	}

	return outBuf.Bytes(), nil
}

func Gzip(input []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	num, err := gz.Write(input)
	if err != nil {
		return nil, err
	}

	if num == 0 || num != len(input) {
		return nil, errors.New("gzip wrote zero or wrong num of bytes")
	}

	err = gz.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func UnGzip(input []byte) ([]byte, error) {
	buf := bytes.NewBuffer(input)

	rdr, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}

	var resBuff bytes.Buffer

	num, err := resBuff.ReadFrom(rdr)
	if err != nil {
		return nil, err
	}

	if num == 0 {
		return nil, errors.New("ungzip read zero bytes")
	}

	return resBuff.Bytes(), nil
}
