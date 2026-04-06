package mcp

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"time"
)

const DEBUG_DIR = "debug_mcp"

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

func recordRemoteReq(body []byte) {
	if !debugEnabled {
		return
	}
	os.MkdirAll(DEBUG_DIR, 0755)
	timestamp := time.Now().Format("20060102150405.000")
	reqFile, _ := os.OpenFile(filepath.Join(DEBUG_DIR, timestamp+"_remote_req.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	defer reqFile.Close()
	io.Copy(reqFile, bytes.NewReader(body))
}

func recordRemoteResp(body io.Reader) io.Reader {
	if !debugEnabled {
		return body
	}
	os.MkdirAll(DEBUG_DIR, 0755)
	timestamp := time.Now().Format("20060102150405.000")
	respFile, _ := os.OpenFile(filepath.Join(DEBUG_DIR, timestamp+"_remote_resp.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	return &debugReader{src: body, file: respFile}
}

func recordLocalStream(stdin io.Writer, stdout io.Reader) (io.Writer, io.Reader) {
	if !debugEnabled {
		return stdin, stdout
	}
 	os.MkdirAll(DEBUG_DIR, 0755)
	timestamp := time.Now().Format("20060102150405.000")
	stdinLogFile, _ := os.OpenFile(filepath.Join(DEBUG_DIR, timestamp+"_local_stdin.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	stdoutLogFile, _ := os.OpenFile(filepath.Join(DEBUG_DIR, timestamp+"_local_stdout.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	return &debugWriter{dst: stdin, file: stdinLogFile}, &debugReader{src: stdout, file: stdoutLogFile}
}
