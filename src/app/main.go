package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
)

type config struct {
	listen string
	dest   string
	log    string
}

func parseArgs(args []string) config {
	if len(args) == 1 {
		log.Println("Missing Parameter: --dest|-d and --url|-u")
		os.Exit(22)
	}
	r := regexp.MustCompile("^(http://)?([A-Za-z0-9.]+:[0-9]{4,})$")
	var out config
	for i := 1; i < len(args); {
		switch args[i] {
		case "--dest", "-d":
			if r.MatchString(args[i+1]) {
				out.dest = args[i+1]
			}
		case "--url", "-u":
			if r.MatchString(args[i+1]) {
				out.listen = args[i+1]
			}
		case "--log", "-l":
			out.log = args[i+1]
		default:
			panic(os.ErrInvalid)
		}
		i = i + 2
	}
	return out
}

func fileWriter(filename string) (*os.File, error) {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) || info.IsDir() {
		return os.Create(filename)
	}
	return os.Open(filename)
}

func makeLogging(log string) (io.Writer, error) {
	if log == "" || log == "stdout" {
		return os.Stdout, nil
	}
	return fileWriter(log)
}

func main() {
	cfg := parseArgs(os.Args)
	writer, err := makeLogging(cfg.log)
	if err != nil {
		panic(err)
	}
	log.SetOutput(writer)
	log.SetPrefix("[ReverseProxy]\t")
	log.SetFlags(log.LstdFlags)

	remote, err := url.Parse(cfg.dest)
	if err != nil {
		panic(err)
	}

	log.Printf("Forwarding to %s", cfg.dest)
	proxy := httputil.NewSingleHostReverseProxy(remote)

	http.HandleFunc("/", handler(proxy))

	log.Printf("Listening on %s", cfg.listen)
	err = http.ListenAndServe(cfg.listen, nil)
	if err != nil {
		panic(err)
	}

}

func handler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s requested %s", r.RemoteAddr, r.URL)
		w.Header().Set("X-Forwarded-For", r.RemoteAddr)
		p.ServeHTTP(w, r)
	}
}
