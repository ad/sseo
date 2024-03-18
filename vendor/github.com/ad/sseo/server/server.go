package server

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ad/sseo/checker"
	"github.com/ad/sseo/files"
)

type Listener struct {
	Server  *http.Server
	Checker *checker.Checker
}

func InitListener(urlChecker *checker.Checker) (*Listener, error) {
	listener := &Listener{
		Checker: urlChecker,
	}

	fs := http.FileServer(http.FS(files.Static))

	mx := http.NewServeMux()
	mx.HandleFunc("/", listener.serveTemplate)
	mx.HandleFunc("/check", listener.checkURL)
	mx.Handle("/static/", neuter(fs))

	s := &http.Server{
		Addr:    ":3301",
		Handler: mx,
	}

	_, cancelCtx := context.WithCancel(context.Background())
	go func(*http.Server) {
		err := s.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Println("server closed")
		} else if err != nil {
			fmt.Printf("error listening for server: %s", err)
		}

		cancelCtx()
	}(s)

	listener.Server = s

	return listener, nil
}

func (l *Listener) serveTemplate(w http.ResponseWriter, r *http.Request) {
	// fmt.Printf("%+v\n", r)

	if r.URL.Path == "" || r.URL.Path == "/" || strings.HasPrefix(r.URL.Path, "/?") {
		r.URL.Path = "/index.html"
	}

	lp := filepath.Join("templates", "layout.html")
	fp := filepath.Join("templates", filepath.Clean(r.URL.Path))

	tmpl, err := template.New("base.html").ParseFS(files.Templates, lp, fp)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, http.StatusText(404), 404)
		return
	}

	err = tmpl.ExecuteTemplate(w, "layout", nil)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, http.StatusText(404), 404)
	}
}

func SubAndWrapFS(fSys fs.FS, dir string) http.FileSystem {
	fSys, err := fs.Sub(fSys, dir)
	if err != nil {
		log.Fatal(err)
	}

	return http.FS(fSys)
}

func neuter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			// fmt.Println("404", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (l *Listener) checkURL(w http.ResponseWriter, r *http.Request) {
	url := strings.Trim(r.FormValue("url"), "\n\r\t ")
	force := r.FormValue("force")

	if url == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	resultProcessing := `
    <div id="result">
            <article>
                <header>
                    <p>
                        <strong>🗓️ Result</strong>
                    </p>
                </header>
             <span aria-busy="true">Processing...</span>
			 <form action="/check" method="post">
				<input type="hidden" name="url" value="` + url + `"  hx-post="/check" hx-swap="outerHTML" hx-target="#result" hx-trigger="every 1s" />
			 </form>
            </article>
    </div>`

	if r, ok := l.Checker.LRU.Get(url); ok {
		if r.Status == "processing" {
			w.Write([]byte(resultProcessing))

			return
		}

		if force != "1" {
			resultCached := `
			<div id="result">
					<article>
						<header>
							<p>
								<strong>🗓️ Result</strong>
							</p>
						</header>
						<p>
						After checking the page, we found the following errors:
					</p>
					<ul>
						<li>` + strings.Join(r.Checks, "</li><li>") + `</li>
					</ul>
					</article>
			</div>`
			w.Write([]byte(resultCached))

			return
		}
	}

	l.Checker.Tasks <- checker.Task{
		URL:   url,
		Force: force == "1",
	}

	w.Write([]byte(resultProcessing))
}
