package provider

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const DEBUG_DIR = "debug_model"

var debugEnabled = false

func SetDebugMode(enabled bool) {
	debugEnabled = enabled
}

type debugReader struct {
	src  io.Reader
	file *os.File
}

func (r *debugReader) Read(p []byte) (len int, err error) {
	len, err = r.src.Read(p)
	if len > 0 {
		r.file.Write(p)
	}
	return len, err
}

type debugWriter struct {
	dst  io.Writer
	file *os.File
}

func (w *debugWriter) Write(p []byte) (n int, err error) {
	w.file.Write(p)
	return w.dst.Write(p)
}

func recordBody(method string, bodyType string, body []byte) {
	if !debugEnabled {
		return
	}
	os.MkdirAll(DEBUG_DIR, 0755)
	timestamp := time.Now().Format("20060102150405.000")
	inFile, _ := os.OpenFile(filepath.Join(DEBUG_DIR, timestamp+"_"+method+"_req_"+bodyType+".log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	defer inFile.Close()
	io.Copy(inFile, bytes.NewReader(body))
}

func recordError(method string, httpCode int, body []byte) {
	if !debugEnabled {
		return
	}
	os.MkdirAll(DEBUG_DIR, 0755)
	timestamp := time.Now().Format("20060102150405.000")
	inFile, _ := os.OpenFile(filepath.Join(DEBUG_DIR, timestamp+"_"+method+"_error_"+strconv.Itoa(httpCode)+".log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	defer inFile.Close()
	io.Copy(inFile, bytes.NewReader(body))
}

func recordStream(method string, src io.Reader, dst io.Writer) (io.Reader, io.Writer) {
	if !debugEnabled {
		return src, dst
	}
	os.MkdirAll(DEBUG_DIR, 0755)
	timestamp := time.Now().Format("20060102150405.000")
	rawLogFile, _ := os.OpenFile(filepath.Join(DEBUG_DIR, timestamp+"_"+method+"_stream_raw.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	convertedLogFile, _ := os.OpenFile(filepath.Join(DEBUG_DIR, timestamp+"_"+method+"_stream_converted.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	return &debugReader{src: src, file: rawLogFile}, &debugWriter{dst: dst, file: convertedLogFile}
}
