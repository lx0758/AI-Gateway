package provider

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

const DEBUG = false
const DEBUG_DIR = "debug"

type debugReader struct {
	src io.Reader
	buf bytes.Buffer
	mu  sync.Mutex
	out *os.File
}

func (r *debugReader) Read(p []byte) (n int, err error) {
	n, err = r.src.Read(p)
	if n > 0 {
		r.mu.Lock()
		r.buf.Write(p[:n])
		for {
			idx := bytes.IndexByte(r.buf.Bytes(), '\n')
			if idx == -1 {
				break
			}
			line := r.buf.Next(idx + 1)
			r.out.Write(line)
		}
		r.mu.Unlock()
	}
	return
}

type debugWriter struct {
	dst io.Writer
	buf bytes.Buffer
	mu  sync.Mutex
	out *os.File
}

func (w *debugWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	w.buf.Write(p)
	for {
		idx := bytes.IndexByte(w.buf.Bytes(), '\n')
		if idx == -1 {
			break
		}
		line := w.buf.Next(idx + 1)
		w.out.Write(line)
	}
	w.mu.Unlock()
	return w.dst.Write(p)
}

func recordBody(method string, bodyType string, body []byte) {
	if !DEBUG {
		return
	}
	os.MkdirAll(DEBUG_DIR, 0755)
	timestamp := time.Now().Format("20060102150405.000")
	inFile, _ := os.OpenFile(filepath.Join(DEBUG_DIR, timestamp+"_"+method+"_req_"+bodyType+".log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	defer inFile.Close()
	io.Copy(inFile, bytes.NewReader(body))
}

func recordError(method string, httpCode int, body []byte) {
	if !DEBUG {
		return
	}
	os.MkdirAll(DEBUG_DIR, 0755)
	timestamp := time.Now().Format("20060102150405.000")
	inFile, _ := os.OpenFile(filepath.Join(DEBUG_DIR, timestamp+"_"+method+"_error_"+strconv.Itoa(httpCode)+".log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	defer inFile.Close()
	io.Copy(inFile, bytes.NewReader(body))
}

func recordStream(method string, src io.Reader, dst io.Writer) (io.Reader, io.Writer) {
	if !DEBUG {
		return src, dst
	}
	os.MkdirAll(DEBUG_DIR, 0755)
	timestamp := time.Now().Format("20060102150405.000")
	inFile, _ := os.OpenFile(filepath.Join(DEBUG_DIR, timestamp+"_"+method+"_stream_raw.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	outFile, _ := os.OpenFile(filepath.Join(DEBUG_DIR, timestamp+"_"+method+"_stream_converted.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	return &debugReader{src: src, out: inFile}, &debugWriter{dst: dst, out: outFile}
}
