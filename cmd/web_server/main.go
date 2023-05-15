
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
	"net/http/httputil"
	"html/template"
	"sort"
	"strconv"
	"bytes"
	"strings"
	"errors"
	"os/exec"
	"encoding/json"
	"time"
	"github.com/qor/render"
	"github.com/google/uuid"
	"gocv.io/x/gocv"	
)

var renderer *render.Render;
var UUIDRegex = regexp.MustCompile(`(\w{8}-\w{4}-\w{4}-\w{4}-\w{12})`)

type Result struct {
	UUID        string   `json:"uuid"`
	PrettyTime  string   `json:"createdAt"`
	ModTime  time.Time
}

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

func getRecentResults() (*[]Result, error) {
	var buf bytes.Buffer
	cmd := exec.Command("find", "./file_storage", "-mtime", "-32", "-iname", "*.json")
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	var err error
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	
	err = cmd.Wait()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			return nil, errors.New(fmt.Sprintf("Recent requests exit code: %d", exiterr.ExitCode()))
		} else {
			return nil, err
		}
	}
	
	var results []Result
	var recentOutput []byte
	recentOutput, err = io.ReadAll(&buf)
	if err != nil {
		return nil, err
	}
	
	lines := strings.Split(string(recentOutput), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		matches := UUIDRegex.FindStringSubmatch(line)
		if matches == nil {
			log.Fatalf("Error when parsing line of recent requests, line was: %s", line)
		}

		var fileInfo os.FileInfo
		fileInfo, err = os.Stat("./file_storage/" + matches[1] + "/data.json")
		if err != nil {
			log.Fatalf("Error when getting file stat of data.json, UUID was %s", matches[1])
		}

		results = append(results, Result{UUID: matches[1], ModTime: fileInfo.ModTime()})
	}

	for i := range results {
		result := &results[i]

		result.PrettyTime = result.ModTime.Format(pkg.TimeLayout)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].ModTime.After(results[j].ModTime)
	})
	

	return &results, nil
}

func writeError(w http.ResponseWriter, msg string, errorNum int) {
	http.Error(w, msg, errorNum)
}

func writeInteralServerError(w http.ResponseWriter, msg string) {
	fmt.Printf("Internal Server Error: %s\n", msg)
	writeError(w, msg, http.StatusInternalServerError)
}

func root(w http.ResponseWriter, req *http.Request) {
	ctx := defaultCtx()

	var err error
	
	var results *[]Result
	results, err = getRecentResults()
	if err != nil {
		writeInteralServerError(w, fmt.Sprintf("Error when retrieving recent requests: %s", err))
		return
	}
	
	var resultBytes []byte
	resultBytes, err = json.Marshal(&results)
	if err != nil {
		writeInteralServerError(w, fmt.Sprintf("Error when marshalling result: %s", err))
		return
	}

	ctx["results"] = string(resultBytes)
	
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
		writeInteralServerError(w, fmt.Sprintf("unable to parse request id"))
		return
	}

	uuid := matches[1]
	ctx["id"] = uuid

	var file *os.File
	var err error
	file, err = os.Open("./file_storage/" + uuid + "/data.json")
	if err != nil {
		writeInteralServerError(w, fmt.Sprintf("Error when reading data json file: %s", err))
		return
	}

	var dataJsonBytes []byte
	dataJsonBytes, err = io.ReadAll(file)
	if err != nil {
		writeInteralServerError(w, fmt.Sprintf("Error when reading data file: %s", err))
		return
	}

	findResult := pkg.FindResult{}
	err = json.Unmarshal(dataJsonBytes, &findResult)
	if err != nil {
		writeInteralServerError(w, fmt.Sprintf("Error when unmarshaling data json: %s", err))
		return
	}

	ctx["found"] = strconv.FormatBool(findResult.Found)
	
	renderer.Execute("request", ctx, req, w)		
}

func makeRequest(w http.ResponseWriter, req *http.Request) {
	var err error
	
	req.ParseMultipartForm(32 << 20)

	if pkg.Debug {
		var dumpBytes []byte
		dumpBytes, err = httputil.DumpRequest(req, false)
		fmt.Printf("\nRequest in Store File:\n%s", string(dumpBytes))		
	}
	
	myUUID := uuid.New()

	var haystackFilename string
	var needleFilename string

	needleFilename, err = storeFile(myUUID, "needleFile", "needle", req)
	if err != nil {
		writeInteralServerError(w, fmt.Sprintf("unable to read needle file: %s", err))
		return
	}
	
	haystackFilename, err = storeFile(myUUID, "haystackFile", "haystack", req)
	if err != nil {
		writeInteralServerError(w, fmt.Sprintf("unable to read haystack file: %s", err))
		return
	}


	findResult := pkg.FindNeedle(haystackFilename, needleFilename)
	defer findResult.Mat.Close()

	gocv.IMWrite("./file_storage/" + myUUID.String() + "/result.png", findResult.Mat)

	var dataFileBytes []byte
	dataFileBytes, err = json.Marshal(findResult)
	if err != nil {
		writeInteralServerError(w, fmt.Sprintf("Error when marshalling data json: %s", err))
		return
	}

	os.WriteFile("./file_storage/" + myUUID.String() + "/data.json", dataFileBytes, 0644)
	fmt.Printf("Wrote data file for request %s\n", myUUID.String())

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

	fs := http.FileServer(http.Dir("./file_storage"))
	http.Handle("/static/", http.StripPrefix("/static", fs))

	assetsFs := http.FileServer(http.Dir("./web_assets"))
	http.Handle("/assets/", http.StripPrefix("/assets", assetsFs))
	
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
