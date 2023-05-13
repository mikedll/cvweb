
package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"io"
	"log"
	"pkg"
	"regexp"
	"mime/multipart"
	"github.com/qor/render"
	"html/template"
	"strings"
	"github.com/google/uuid"
	"gocv.io/x/gocv"
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

func storeFile(prefix uuid.UUID, param string, serverName string, req *http.Request) (string, error) {
	var err error

	withUUID := "./file_storage/" + prefix.String()
	err = os.MkdirAll(withUUID, os.ModePerm)
	if err != nil {
		return "", err
	}
	
	var file multipart.File
	var header *multipart.FileHeader
	file, header, err = req.FormFile(param)
	if err != nil {
		return "", err
	}

	var fileBytes []byte
	fileBytes, err = io.ReadAll(file)
	if err != nil {
		return "", err
	}
	
	localFilename := withUUID + "/" + serverName + path.Ext(header.Filename)
	err = os.WriteFile(localFilename, fileBytes, 0644)
	if err != nil {
		return "", err
	}

	return localFilename, nil
}

func request(w http.ResponseWriter, req *http.Request) {
	ctx := defaultCtx()

	myRegex := regexp.MustCompile(`/requests/([a-z0-9-]+)`)
	matches := myRegex.FindStringSubmatch(req.URL.Path)
	if matches == nil {
		writeInteralServerError(w, fmt.Sprintf("unable to parse request id: %s"))
		return
	}
	
	ctx["id"] = matches[1]
	renderer.Execute("request", ctx, req, w)		
}

func makeRequest(w http.ResponseWriter, req *http.Request) {
	req.ParseMultipartForm(32 << 20)

	myUUID := uuid.New()

	var err error
	var haystackFilename string
	var needleFilename string
	haystackFilename, err = storeFile(myUUID, "haystackFile", "haystack", req)
	if err != nil {
		writeInteralServerError(w, fmt.Sprintf("unable to read haystack file: %s", err))
		return
	}

	needleFilename, err = storeFile(myUUID, "needleFile", "needle", req)
	if err != nil {
		writeInteralServerError(w, fmt.Sprintf("unable to read needle file: %s", err))
		return
	}

	haystackWithBox := pkg.FindNeedle(haystackFilename, needleFilename)
	defer haystackWithBox.Close()

	gocv.IMWrite("./file_storage/" + myUUID.String() + "/result.png", haystackWithBox)

	http.Redirect(w, req, req.URL.Host + "/requests/" + myUUID.String(), 302)
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
	http.Handle("/requests/", http.HandlerFunc(request))
	http.Handle("/requests", http.HandlerFunc(makeRequest))
	
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
