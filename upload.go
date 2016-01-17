package main

import (
    "html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"fmt"
	//"io/ioutil"
)

//Compile templates on start
var templates = template.Must(template.ParseFiles("tmpl/upload.html"))

//Display the named template
func display(w http.ResponseWriter, tmpl string, data interface{}) {
	templates.ExecuteTemplate(w, tmpl+".html", data)
}

func ConvertOfficeDocToPdf(fileIn string, fileOut string, port int) {
	args := []string{"-f", "pdf",
		"-eSelectPdfVersion=1",
		"-eReduceImageResolution=true",
		"-eMaxImageResolution=300",
		"-p",
		strconv.Itoa(port),
		"-o",
		fileOut,
		fileIn,
	}
	path, err := exec.LookPath("unoconv")
	if err != nil {
		fmt.Printf("Cannot find unoconv in PATH")
	}
	fmt.Printf("unoconv is available at %s\n", path)
	cmd := exec.Command("unoconv", args...)
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("Error: ", err)
	} else {
		fmt.Printf("Success: %s\n", out)
	}
}

//This is where the action happens.
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	//GET displays the upload form.
	case "GET":
		display(w, "upload", nil)

	//POST takes the uploaded file(s) and saves it to disk.
	case "POST":
		//get the multipart reader for the request.
		reader, err := r.MultipartReader()
		var outFile string

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		//copy each part to destination.
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}

			//if part.FileName() is empty, skip this iteration.
			if part.FileName() == "" {
				continue
			}
			outFile = "tmp/" + part.FileName()
			dst, err := os.Create("tmp/" + part.FileName())
			defer dst.Close()

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			
			if _, err := io.Copy(dst, part); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		ConvertOfficeDocToPdf(outFile, "tmp/foo.pdf", 8100)
		//dat, err := ioutil.ReadFile("tmp/foo.pdf")

		http.ServeFile(w, r, "tmp/foo.pdf")

		//display success message.
		//display(w, "upload", "Upload successful.")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/upload", uploadHandler)

	//static file handler.
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	//Listen on port 8080
	http.ListenAndServe(":8088", nil)
}


