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

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", root)
	mux.HandleFunc("/api/extract", extract)
	server := &http.Server{
		Addr: "127.0.0.1:8080",
		Handler: mux,
	}
	fmt.Println("Server started.")
	server.ListenAndServe()
}

func root(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("./templates/root.html")
	t.Execute(w, nil)
}

func extract(w http.ResponseWriter, r *http.Request) {
	fmt.Println("\nReq sent to extract.")
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
	uploadedImage, _, err := image.Decode(file)
	if err != nil {
		fmt.Println("Error decoding file.")
		fmt.Println(err)
		return
	}

	numOfColors, err := strconv.Atoi(r.FormValue("numOfColors"))
	if err != nil {
		fmt.Println("Error retrieving numOfColors.")
		fmt.Println(err)
		return
	}

	colors := imageManip.ExtractPalette(uploadedImage, numOfColors)
	fmt.Println("Colors:", colors)

	ret, err := json.Marshal(colors)
	if err != nil {
		fmt.Println("Error marshalling json.")
		fmt.Println(err)
		return
	}

	fmt.Println("Sending json:", string(ret))
	w.Header().Set("Content-Type", "application/json")
	w.Write(ret)
	return
}
