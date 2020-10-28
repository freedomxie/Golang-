package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var proxyMap map[string]string
var port string
var path string
var ssl string
var key string
var crt string
var filePath string

const (
	ListDir = 0x0001
)

func isExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return os.IsExist(err)
}

func staticDirHandler(mux *http.ServeMux, prefix string, staticDir string, flags int) {
	mux.HandleFunc(prefix, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Action, Module")

		remouteURL := ""
		for k, v := range proxyMap {
			if strings.Contains(r.RequestURI, k) {
				remouteURL = v
				break
			}
		}

		if remouteURL != "" {
			log.Println("route:", r.RequestURI, "  => ", remouteURL+r.RequestURI)
			remote, err := url.Parse(remouteURL)
			if err != nil {
				panic(err)
			}
			proxy := httputil.NewSingleHostReverseProxy(remote)
			r.URL.Host = remote.Host
			r.URL.Scheme = remote.Scheme
			// r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
			r.Host = remote.Host
			proxy.ServeHTTP(w, r)
		} else {
			file := staticDir + r.URL.Path[len(prefix)-1:]
			if (flags & ListDir) == 0 {
				if exists := isExists(file); !exists {
					http.NotFound(w, r)
					return
				}
			}
			http.ServeFile(w, r, file)
		}
	})
}

func initConfig(configFile string) {
	log.Println("read config file now ...")
	f, err := os.Open(configFile)
	if err != nil {
		log.Println("Failed to read the configuration file: conf.ini")
		os.Exit(1)
	}
	defer f.Close()

	br := bufio.NewReader(f)

	proxyMap = make(map[string]string)

	for {
		line, err := br.ReadString('\n')

		if err == io.EOF {
			break
		}

		if 0 == len(line) || line == "\n" {
			continue
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "port") {
			port = strings.Trim(strings.Split(line, "=")[1], " ")
			port = strings.Replace(port, "\n", "", -1)
			continue
		}

		if strings.HasPrefix(line, "ssl") {
			ssl = strings.Trim(strings.Split(line, "=")[1], " ")
			ssl = strings.Replace(ssl, "\n", "", -1)
			continue
		}

		if strings.HasPrefix(line, "key") {
			key = strings.Trim(strings.Split(line, "=")[1], " ")
			key = strings.Replace(key, "\n", "", -1)
			continue
		}

		if strings.HasPrefix(line, "crt") {
			crt = strings.Trim(strings.Split(line, "=")[1], " ")
			crt = strings.Replace(crt, "\n", "", -1)
			continue
		}

		if strings.HasPrefix(line, "path") {
			path = strings.Trim(strings.Split(line, "=")[1], " ")
			path = strings.Replace(path, "\n", "", -1)
			continue
		}

		if strings.HasPrefix(line, "filePath") {
			filePath = strings.Trim(strings.Split(line, "=")[1], " ")
			filePath = strings.Replace(filePath, "\n", "", -1)
			continue
		}

		line = strings.Replace(line, "\n", "", -1)

		if strings.HasSuffix(line, "/") {
			log.Println("config is error,end by /")
			os.Exit(0)
		}

		if strings.HasSuffix(line, "\\") {
			log.Println("config is error,end by \\")
			os.Exit(0)
		}

		arr := strings.Split(line, "=")
		keyArr := strings.Split(arr[0], ",")

		v := strings.Replace(arr[1], "\n", "", -1)

		for i := 0; i < len(keyArr); i++ {
			proxyMap[keyArr[i]] = v
			log.Println("input map:", keyArr[i], v)
		}
		continue
	}
}

func debug() {
	fmt.Println("port:", port)
	fmt.Println("ssl:", ssl)
	fmt.Println("key:", key)
	fmt.Println("crt:", crt)
	fmt.Println("filePath:", filePath)
}

const tpl = `<html>
<head>
<title>上传文件</title>
</head>
<body>
<form enctype="multipart/form-data" action="/upload" method="post">
 <input type="file" name="file" />
 <input type="submit" value="上传" />
</form>
</body>
</html>`

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(tpl))
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {

	//body, _ := ioutil.ReadAll(r.Body) //把  body 内容读入字符串 s
	nowtime := time.Now()
	switch r.Method {
	case "POST":
		err := r.ParseMultipartForm(10 * 1000 * 1000)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		m := r.MultipartForm
		files := m.File["file"]
		for i, _ := range files {
			file, err := files[i].Open()
			defer file.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			dst, err := os.Create(filePath + "/" + files[i].Filename)
			defer dst.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if _, err := io.Copy(dst, file); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	elapsed := time.Since(nowtime)
	fmt.Println("App elapsed: ", elapsed)
	result := strconv.FormatInt(elapsed.Milliseconds(), 10)
	w.Write([]byte(result + " ms"))

}

func main() {
	argNum := len(os.Args)
	if argNum != 2 {
		fmt.Println("useAge: nohup ./WebProxy config.ini  &")
		os.Exit(0)
	}
	configFile := os.Args[1]
	initConfig(configFile)
	debug()

	mux := http.NewServeMux()
	staticDirHandler(mux, "/", path, 0)
	mux.HandleFunc("/upload", uploadHandler)
	mux.HandleFunc("/index", indexHandler)
	var err error
	if ssl == "false" {
		err = http.ListenAndServe(":"+port, mux)
	} else {
		err = http.ListenAndServeTLS(":"+port, crt, key, mux)
	}
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}
