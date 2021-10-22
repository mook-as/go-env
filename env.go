// -*- go -*-
package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"sort"
	"strings"
)

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	// - /static/ws.html
	// - /static/js/jquery-2.1.4.min.js
	http.Handle("/fs/", http.StripPrefix("/fs/", http.HandlerFunc(fsHandler)))
	http.HandleFunc("/", envHandler)
	http.HandleFunc("/body", bodyHandler)
	http.HandleFunc("/crash", crashHandler)
	http.HandleFunc("/headers", headersHandler)
	http.HandleFunc("/memory/consume", consumeMemoryHandler)
	http.HandleFunc("/memory/release", releaseMemoryHandler)
	http.HandleFunc("/memory", memoryStatsHandler)
	http.HandleFunc("/stdout", stdoutHandler)
	addr := ":" + os.Getenv("PORT")
	fmt.Printf("Listening on %v\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// envHandler prints out the environment seen by the backend/application/this process
func envHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, strings.Join(os.Environ(), "\n"))
}

// crashHandler kills the application
func crashHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Crashing...")
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
	os.Exit(1)
}

// headerHandler prints out the active headers in the request
func headersHandler(w http.ResponseWriter, req *http.Request) {
	req.Header.Write(w)
}

func stdoutHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(os.Stdout, "Message on stdout")
	fmt.Fprintln(os.Stderr, "Message on stderr")
	fmt.Fprintln(w, "Message written")
}

type memorySize struct {
	size uint64
}

func (s memorySize) String() string {
	remaining := float64(s.size)
	units := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}
	for len(units) > 1 && remaining > 512 {
		remaining /= 1024
		units = units[1:]
	}
	return fmt.Sprintf("%0.2f %s", remaining, units[0])
}

func getMemStats() string {
	memStats := runtime.MemStats{}
	runtime.ReadMemStats(&memStats)
	return fmt.Sprintf("Allocated %v (%v heap), %v from system (%v heap)",
		memorySize{memStats.Alloc}, memorySize{memStats.HeapAlloc},
		memorySize{memStats.Sys}, memorySize{memStats.HeapSys})
}

var buf []interface{}

func consumeMemoryHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Consuming memory...\n")
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	for i := 0; i < 99; i++ {
		if flusher, ok := w.(http.Flusher); ok {
			fmt.Fprintf(w, "%s\n", getMemStats())
			flusher.Flush()
		}
		buf = append(buf, bytes.Repeat([]byte{byte(i)}, 1024*1024))
	}

	fmt.Fprintf(w, "Consumption complete\n")
	fmt.Fprintf(w, "%s\n", getMemStats())
	fmt.Fprintf(w, "Running GC...\n")
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	runtime.GC()
	fmt.Fprintf(w, "%s\n", getMemStats())
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func releaseMemoryHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Releasing memory...\n")
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
	for len(buf) > 0 {
		if flusher, ok := w.(http.Flusher); ok {
			fmt.Fprintf(w, "%s\n", getMemStats())
			flusher.Flush()
		}
		buf = buf[1:]
	}
	fmt.Fprintf(w, "%s\n", getMemStats())
	fmt.Fprintf(w, "Running GC...\n")
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
	runtime.GC()
	fmt.Fprintf(w, "%s\n", getMemStats())
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func memoryStatsHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "%s\n", getMemStats())
}

func bodyHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	w.Header().Set("Content-Type", "text/plain")

	outFile, err := os.Create("/tmp/last-access.dat")
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer outFile.Close()
	for k, v := range req.Header {
		fmt.Fprintf(outFile, "%s: %s\r\n", k, v)
	}
	fmt.Fprintf(w, "\r\n")

	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, req.Body); err != nil {
		fmt.Printf("Error buffering data: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, err := outFile.Write(buf.Bytes()); err != nil {
		fmt.Printf("Error saving data: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(buf.Bytes()); err != nil {
		fmt.Printf("Error dumping data: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "\n\nError dumping data: %v\n", err)
		return
	}
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func fsHandler(w http.ResponseWriter, req *http.Request) {
	absPath := path.Join("/", req.URL.Path)
	f, err := os.Open(absPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	if info.IsDir() {
		fmt.Fprintf(w, "Path %s is a directory:\n", absPath)
		children, err := f.Readdir(0)
		sort.SliceStable(children, func(i, j int) bool {
			return children[i].Name() < children[j].Name()
		})
		for _, child := range children {
			line := fmt.Sprintf("%-10s %8d %10s %s", child.Name(), child.Size(), child.Mode(), child.ModTime())
			if child.Mode() & os.ModeSymlink != 0 {
				target, e := os.Readlink(path.Join(absPath, child.Name()))
				if e != nil {
					fmt.Fprintf(w, "%s <error reading link: %s>\n", line, err)
				} else {
					fmt.Fprintf(w, "%s -> %s\n", line, target)
				}
			} else {
				fmt.Fprintln(w, line)
			}
		}
		if err != nil {
			w.Write([]byte(err.Error()))
		}
	} else {
		_, err = io.Copy(w, f)
		if err != nil {
			w.Write([]byte(err.Error()))
		}
	}
}