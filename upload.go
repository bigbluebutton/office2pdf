package main

import (
    "html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"fmt"
	"strings"
	"path/filepath"
//	"bytes"
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

func ExecuteCommand(app string, args []string) {
	cmd := exec.Command(app, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
	    fmt.Println(fmt.Sprint(err) + ": " + string(output))
	    return
	} else {
	    fmt.Println(string(output))
	}

}

func WindowsConvertOfficeDocToPdf(fileIn string, fileOut string) {
	args := []string{
		"/verbose",
		fileIn,
		fileOut,
	}
	path, err := exec.LookPath("OfficeToPDF.exe")
	if err != nil {
		fmt.Printf("Cannot find OfficeToPDF.exe in PATH")
	}
	fmt.Printf("OfficeToPDF.exe is available at %s\n", path)
	ExecuteCommand("OfficeToPDF.exe", args)
//	cmd := exec.Command("OfficeToPDF.exe", args...)
//	out, err := cmd.Output()
//	if err != nil {
//		fmt.Printf("Error: ", err)
//	} else {
//		fmt.Printf("Success: %s\n", out)
//	}
	
//	var out bytes.Buffer
//	var stderr bytes.Buffer
//	cmd.Stdout = &out
//	cmd.Stderr = &stderr
//	err2 := cmd.Run()
//	if err2 != nil {
//    	fmt.Println(fmt.Sprint(err2) + ": " + stderr.String())
//    	return
//	}
//	fmt.Println("Result: " + out.String())
}

// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherise, attempt to create a hard link
// between the two files. If that fail, copy the file contents from src to dst.
func CopyFile(src, dst string) (err error) {
    sfi, err := os.Stat(src)
    if err != nil {
        return
    }
    if !sfi.Mode().IsRegular() {
        // cannot copy non-regular files (e.g., directories,
        // symlinks, devices, etc.)
        return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
    }
    dfi, err := os.Stat(dst)
    if err != nil {
        if !os.IsNotExist(err) {
            return
        }
    } else {
        if !(dfi.Mode().IsRegular()) {
            return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
        }
        if os.SameFile(sfi, dfi) {
            return
        }
    }
    if err = os.Link(src, dst); err == nil {
        return
    }
    err = copyFileContents(src, dst)
    return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
    in, err := os.Open(src)
    if err != nil {
        return
    }
    defer in.Close()
    out, err := os.Create(dst)
    if err != nil {
        return
    }
    defer func() {
        cerr := out.Close()
        if err == nil {
            err = cerr
        }
    }()
    if _, err = io.Copy(out, in); err != nil {
        return
    }
    err = out.Sync()
    return
}

func SaveFile(w http.ResponseWriter, r *http.Request) (string, error) {
	fmt.Printf("Saving file \n")
		reader, err := r.MultipartReader()
		var outFile string

		if err != nil {
			fmt.Printf("Cannot get MultipartReader\n")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return "", err
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
			fmt.Printf("Saving file " + part.FileName() + "\n")
			outFile = "tmp/" + part.FileName()
			dst, err := os.Create("tmp/" + part.FileName())
			defer dst.Close()

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return "", err
			}
			
			if _, err := io.Copy(dst, part); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return "", err
			}
		}

		return outFile, nil
}

//This is where the action happens.
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	//GET displays the upload form.
	case "GET":
		display(w, "upload", nil)

	//POST takes the uploaded file(s) and saves it to disk.
	case "POST":
		fmt.Printf("Handling file upload\n")
		//get the multipart reader for the request.
/*		reader, err := r.MultipartReader()
		var outFile, fileName string

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
			fileName = part.FileName()
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
*/
		outFile, err := SaveFile(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Printf("Upload file is " + outFile + "\n")
		name := strings.TrimSuffix(outFile, filepath.Ext(outFile))
//		dupFile := "tmp/copy-" + fileName
//		CopyFile(outFile, dupFile)
		//ConvertOfficeDocToPdf(outFile, "tmp/foo.pdf", 8100)
		pdfFile := name + ".pdf"
		fmt.Printf("PDF file is " + pdfFile + "\n")
		WindowsConvertOfficeDocToPdf(outFile, pdfFile)
		//WindowsConvertOfficeDocToPdf(outFile, "tmp/foo.pdf")
		//dat, err := ioutil.ReadFile("tmp/foo.pdf")

		http.ServeFile(w, r, pdfFile)

		//display success message.
		display(w, "upload", "Upload successful.")
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


