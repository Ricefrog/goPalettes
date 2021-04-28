package main

import (
	"fmt"
	"net/http"
	"html/template"
	"encoding/json"
	"goPalettes/imageManip"
	"strconv"
	"image"
	_ "image/png"
	_ "image/jpeg"
)

var UPLOADED_IMAGE image.Image
//var COLORS []byte

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/upload", upload)
	mux.HandleFunc("/api/extract/", extract)
	//mux.HandleFunc("/", root)
	//mux.HandleFunc("/api/extracted", extracted)
	server := &http.Server{
		Addr: "127.0.0.1:8080",
		Handler: mux,
	}
	fmt.Println("Server started.")
	server.ListenAndServe()
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func root(w http.ResponseWriter, r *http.Request) {
	fmt.Println("\nRequest sent to root.", r.Method)
	t, _ := template.ParseFiles("./templates/root.html")
	t.Execute(w, nil)
}

// POST form with file so it can be saved in local storage.
func upload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("\nRequest sent to upload.", r.Method)
	// get the content from the POSTed from
	r.ParseMultipartForm(10485760)
	file, _, err := r.FormFile("image")
	if err != nil {
		fmt.Println("Error uploading file.")
		fmt.Println(err)
		return
	}
	defer file.Close()

	// decode file to image
	UPLOADED_IMAGE, _, err = image.Decode(file)
	if err != nil {
		fmt.Println("Error decoding file.")
		fmt.Println(err)
		return
	}

	enableCors(&w)
	w.WriteHeader(200)
	return
}

// GET, include number of colors to extract in the url as a query parameter with the number of colors to extract.
func extract(w http.ResponseWriter, r *http.Request) {
	fmt.Println("\nRequest sent to extract.", r.Method)
	if UPLOADED_IMAGE == nil {
		// return "No Content" header if there is no uploaded image to use.
		w.WriteHeader(204)
		return
	}

	keys, ok := r.URL.Query()["colors"]
	if !ok || len(keys[0]) < 1{
		fmt.Println("URL param \"colors\" is missing. ")
	}

	numOfColors, err := strconv.Atoi(keys[0])
	if err != nil {
		fmt.Println("Invalid number of colors.")
		return
	}

	colors := imageManip.ExtractPalette(UPLOADED_IMAGE, numOfColors)
	fmt.Println("Colors:", colors)

	ret, err := json.Marshal(colors)
	if err != nil {
		fmt.Println("Error marshalling json.")
		fmt.Println(err)
		return
	}

	fmt.Println("Sending json:", string(ret))
	enableCors(&w)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(ret)
	return
}

/*
// testing: uploads json data of colors after a color extraction.
func extracted(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Req to extracted.")
	w.Write(append(COLORS))
	return
}
*/
