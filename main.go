package main

import (
	"bytes"
	"context"
	_ "embed"
	"flag"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"golang.org/x/crypto/acme/autocert"
)

//go:embed pkgs.txt
var pkgs []byte

const (
	hostname = "golang.unexpl0.red"
)

const tplRaw = `<html>
<head>
<title>{{ .PackageName }}: dzm golang repo</title><meta name="go-import" 
content="{{ .HostName }}/{{ .PackageName }} git https://{{ .HostName }}/{{ .PackageName }}">
<meta name="go-source" content="{{ .HostName }}/{{ .PackageName }} _ {{ .URL }}{/dir} {{ .URL }}{/dir}/{file}#n{line}">
</head>
<body>
go get golang.unexpl0.red/{{ .PackageName }}
</body>
</html>`

var (
	email      string
	uid        int
	gid        int
	packageMap map[string]string
	tpl        *template.Template
)

func init() { // flags
	flag.StringVar(&email, email, "", "contact email")
	flag.IntVar(&uid, "uid", 0, "uid to execute with after binding to 443")
	flag.IntVar(&gid, "gid", 0, "gid to execute with after binding to 443")
}

func init() { // package map
	// initialize package name to URL map
	packageMap = make(map[string]string)

	t := bytes.Split(pkgs, []byte("\n"))

	for _, keyvalue := range t {
		kv := bytes.Split(keyvalue, []byte(" "))
		if len(kv) != 2 {
			if len(keyvalue) == 0 {
				log.Println("init():::encountered empty line, returning...")
				return
			}
			panic("Malformed config!")
		}
		key := string(kv[0])
		value := string(kv[1])

		packageMap[key] = value

		log.Printf("init():::added %s -> %s to map", key, value)
	}
}

func init() { // template
	t, err := template.New("output").Parse(tplRaw)
	if err != nil {
		log.Fatal("indexHandler():::error parsing template", err)
	}
	tpl = t
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)

	mgr := &autocert.Manager{
		Cache:      autocert.DirCache("cert-dir"),
		Prompt:     autocert.AcceptTOS,
		Email:      email,
		HostPolicy: autocert.HostWhitelist(hostname),
	}

	server := &http.Server{
		Handler:   mux,
		Addr:      ":https",
		TLSConfig: mgr.TLSConfig(),
	}

	l, err := net.Listen("tcp", ":https")
	if err != nil {
		log.Fatal("main():::Error initializing listener: ", err)
	}

	log.Print("Privileged bind successful, dropping priveleges...")
	// drop priveleges
	if err := syscall.Setgid(gid); err != nil {
		log.Fatal("main():::Error dropping privleges: ", err)
	}

	if err := syscall.Setuid(uid); err != nil {
		log.Fatal("main():::Error dropping privelges: ", err)
	}
	log.Println("Done!")

	interrupts := make(chan os.Signal, 1)
	finished := make(chan interface{}, 1)
	signal.Notify(interrupts, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-interrupts
		log.Print("Interrrupt received! Draining connections and shutting down...")
		if err := server.Shutdown(context.Background()); err != nil {
			panic(err)
		}
		log.Println("done!")
		finished <- "yes"
	}()

	if err := server.ServeTLS(l, "", ""); err != http.ErrServerClosed {
		log.Fatal("main():::fatal error in http server: ", err)
	}

	<-finished
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// no need to sychronize access to this map, as it only is modified on init, all
	// other access is read access
	key := strings.TrimPrefix(r.URL.Path, "/")
	value, exists := packageMap[key]
	if !exists {
		http.NotFound(w, r)
		return
	}

	output := struct {
		PackageName string
		HostName    string
		URL         string
	}{key, hostname, value}

	if err := tpl.Execute(w, output); err != nil {
		log.Println("indexHandler:::failed to execute template: " + err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
