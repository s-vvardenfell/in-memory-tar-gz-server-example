package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"time"
)

type HttpClientOpts struct {
	AddrToUpload string
	Timeout      time.Duration
}

type HttpClient struct {
	client *http.Client
}

func NewHttpClient(opts HttpClientOpts) *HttpClient {
	tr := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   opts.Timeout * time.Second,
			KeepAlive: opts.Timeout * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   opts.Timeout * time.Second,
		ResponseHeaderTimeout: opts.Timeout * time.Second,
		MaxIdleConns:          10,
		MaxIdleConnsPerHost:   5,
		IdleConnTimeout:       opts.Timeout * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}

	return &HttpClient{
		client: &http.Client{
			Transport: tr,
			Timeout:   opts.Timeout * time.Second,
		},
	}
}

// Reading large file by parts
func (cl *HttpClient) UploadFileByParts(filename, addrToUpload string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	var totalBytesRead int
	var part int

	reader := bufio.NewReader(file)

	for {
		buf := make([]byte, MaxBodyBytes)

		bytesRead, err := reader.Read(buf)
		totalBytesRead += bytesRead

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		if bytesRead > 0 { // send part using multipart-upload
			err = cl.multipartUpload(addrToUpload, buf[:bytesRead])
			if err != nil {
				return err
			}
		}

		part++
	}

	fmt.Printf("successfull file upload; total parts: %d, bytes send: %d\n", part, totalBytesRead)
	return nil
}

func (cl *HttpClient) multipartUpload(addr string, msg []byte) error {
	fmt.Printf("1) uploading file size: %d\n", len(msg))

	tarredResp, err := Tar(msg)
	if err != nil {
		return err
	}

	fmt.Printf("2) uploading file size after taring: %d\n", len(tarredResp))

	gzippedResp, err := Gzip(tarredResp)
	if err != nil {
		return err
	}

	fmt.Printf("3) uploading file size after gzipping: %d\n", len(gzippedResp))

	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)

	part, err := writer.CreateFormFile("file", "tmp_archive") // file will be acessed on server by 'file'
	if err != nil {
		return err
	}

	fileReader := bytes.NewReader(gzippedResp)

	num, err := io.Copy(part, fileReader)
	if err != nil || num == 0 {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, addr, buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType()) // important thing

	resp, err := cl.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return err
}
