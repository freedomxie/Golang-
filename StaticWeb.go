package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

var uploadDir string

func uploadHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "POST":
		err := r.ParseMultipartForm(100000)
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
			dst, err := os.Create(uploadDir + "/" + files[i].Filename)
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
}

func main() {

	arg_num := len(os.Args)

	if arg_num != 3 {
		fmt.Println("useAge:")
		fmt.Println("\t run              : ./StaticWeb  path  port")
		fmt.Println("\t path is like     : /home/xx/xx")
		fmt.Println("\t upload file      : curl -F \"file=@/youpath/../image.jpg\" \"http://ip:port/upload\"")
		fmt.Println("\t view upload file : http://ip:port")
		os.Exit(0)
	}
	uploadDir = os.Args[1]
	port := os.Args[2]

	fmt.Println("The Web Path is:", os.Args[1], " port:"+port)
	http.Handle("/", http.FileServer(http.Dir(uploadDir)))
	http.HandleFunc("/upload", uploadHandler)
	http.ListenAndServe(":"+port, nil)
}
