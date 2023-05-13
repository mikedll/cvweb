
package main

import (
	"fmt"
	"net/http"
	"os"
	"io"
	"log"
	"pkg"
	"mime/multipart"
	"github.com/qor/render"
	"html/template"
	"strings"
)

var renderer *render.Render;

func defaultCtx() map[string]interface{} {
	ctx := make(map[string]interface{})
	if pkg.Env == "production" {
		snippet := `
		<!-- Google tag (gtag.js) -->
		<script async src="https://www.googletagmanager.com/gtag/js?id=ID"></script>
		<script>
			window.dataLayer = window.dataLayer || [];
			function gtag(){dataLayer.push(arguments);}
			gtag('js', new Date());

			gtag('config', 'ID');
		</script>
`
		snippet = strings.ReplaceAll(snippet, "ID", os.Getenv("GOOGLE_ANALYTICS_ID"))

		if pkg.Debug {
			fmt.Printf("snippet:\n %s\n", snippet)
		}
		
		ctx["googleAnalytics"] = template.HTML(snippet)
	}
	return ctx
}

func writeError(w http.ResponseWriter, msg string, errorNum int) {
	http.Error(w, msg, errorNum)
}

func writeInteralServerError(w http.ResponseWriter, msg string) {
	writeError(w, msg, http.StatusInternalServerError)
}

func root(w http.ResponseWriter, req *http.Request) {
	ctx := defaultCtx()
	renderer.Execute("index", ctx, req, w)		
}

func storeFile(param string, req *http.Request) error {
	var err error
	var file multipart.File
	var header *multipart.FileHeader
	file, header, err = req.FormFile(param)
	if err != nil {
		return err
	}

	var fileBytes []byte
	fileBytes, err = io.ReadAll(file)
	if err != nil {
		return err
	}
	err = os.WriteFile("./file_storage/" + header.Filename, fileBytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

func run(w http.ResponseWriter, req *http.Request) {
	ctx := defaultCtx()
	req.ParseMultipartForm(32 << 20)

	var err error
	err = storeFile("haystackFile", req)
	if err != nil {
		writeInteralServerError(w, fmt.Sprintf("unable to read haystack file: %s", err))
		return
	}
	
	renderer.Execute("run", ctx, req, w)			
}

func main() {
	pkg.Init()
	
	renderer = render.New(&render.Config{
		ViewPaths:     []string{ "web_app_views" },
		DefaultLayout: "application",
		FuncMapMaker:  nil,
	})

	fmt.Printf("Web server loading for env %s...\n", pkg.Env)

	http.Handle("/", http.HandlerFunc(root))
	http.Handle("/run", http.HandlerFunc(run))
	
	var addr string = "localhost:8081"
	port := os.Getenv("PORT")

	if port != "" {
		addr = fmt.Sprintf("localhost:%s", port)
	}
		
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("Error on ListenAndServe: %s\n", err)
	}
}
