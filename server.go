package main

import (
	"archive/zip"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var apiKey = ""

func getLocalVersion() string { return "0.2" }

func main() {

	keyFile, err := os.ReadFile("./key.txt")
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}
	apiKey = strings.TrimSuffix(string(keyFile), "\n")
	fmt.Println(apiKey)

	mux := http.NewServeMux()
	mux.HandleFunc("/", basicAuth(getRoot))
	mux.HandleFunc("/hello", basicAuth(getHello))
	mux.HandleFunc("/resource_pack", basicAuth(getResourcePack))
	mux.HandleFunc("/outdated", basicAuth(getOutdated))
	mux.HandleFunc("/mod", basicAuth(getModfolderPack))
	mux.HandleFunc("/create/prop", basicAuth(postCreatePropFile))
	mux.HandleFunc("/create/elytra_prop", basicAuth(postCreateElytraPropFile))
	mux.HandleFunc("/create/texture", basicAuth(postCreateTexture))

	http.ListenAndServe(":3333", mux)
}

func zipFolder() {
	// Folder to be zipped
	folderToZip := "template"

	// Name of the zip file
	zipFileName := "resource_pack.zip"

	// Create a new zip file
	zipFile, err := os.Create(zipFileName)
	if err != nil {
		fmt.Println("Error creating zip file:", err)
		return
	}
	defer zipFile.Close()

	// Create a new zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Walk through the folder and its subfolders
	err = filepath.Walk(folderToZip, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path within the folder
		relPath, err := filepath.Rel(folderToZip, path)
		if err != nil {
			return err
		}

		// If it's a directory, create a new directory in the zip file
		if info.IsDir() {
			return nil
		}

		// Create a new file in the zip file
		fileInZip, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		// Open the file to be zipped
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// Copy the file content to the zip file
		_, err = io.Copy(fileInZip, file)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		fmt.Println("Error zipping folder:", err)
		return
	}

	fmt.Println("Folder zipped successfully.")
}

func makePropFile(item string, displayName string, texture string, fileName string) {
	filePath := "template/assets/minecraft/citresewn/cit/" + fileName + ".properties"

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	// Write key-value pairs to the file
	_, err = fmt.Fprintf(file, "%s=%s\n", "type", "item")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	_, err = fmt.Fprintf(file, "%s=%s\n", "items", item)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	_, err = fmt.Fprintf(file, "%s=%s\n", "nbt.display.Name", displayName)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	_, err = fmt.Fprintf(file, "%s=%s", "texture", "textures/"+texture)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("File created:", filePath)
}

func makeElytraPropFile(displayName string, texture string, fileName string) {
	filePath := "template/assets/minecraft/citresewn/cit/" + fileName + ".properties"

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	// Write key-value pairs to the file
	_, err = fmt.Fprintf(file, "%s=%s\n", "type", "elytra")
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	_, err = fmt.Fprintf(file, "%s=%s\n", "nbt.display.Name", displayName)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	_, err = fmt.Fprintf(file, "%s=%s", "texture", "textures/"+texture)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("File created:", filePath)
}

func saveImage(url string, fileName string) {
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer response.Body.Close()

	// Create a new file to save the image
	file, err := os.Create("template/assets/minecraft/citresewn/cit/textures/" + fileName + ".png")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	// Copy the response body to the file
	_, err = io.Copy(file, response.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Image downloaded successfully.")
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	/*hasFirst := r.URL.Query().Has("first")
	first := r.URL.Query().Get("first")
	hasSecond := r.URL.Query().Has("second")
	second := r.URL.Query().Get("second")

	fmt.Printf("got / request. first(%t)=%s, second(%t)=%s\n", hasFirst, first, hasSecond, second)
	io.WriteString(w, "This is my website!\n")*/
	fmt.Printf("Server pinged!\n")
	io.WriteString(w, "Pong")
}

func getHello(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got /hello request\n")
	io.WriteString(w, "Hello, HTTP!\n")
}

func getResourcePack(w http.ResponseWriter, r *http.Request) {
	zipFolder()
	file, err := os.Open("resource_pack.zip")
	if err != nil {
		http.Error(w, "Unable to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Set the Content-Type header to indicate the file type
	w.Header().Set("Content-Type", "application/zip")

	// Send the zip file to the client
	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Unable to send file", http.StatusInternalServerError)
		return
	}
}

func getModfolderPack(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("picklesclientsideutils.jar")
	if err != nil {
		http.Error(w, "Unable to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Set the Content-Type header to indicate the file type
	w.Header().Set("Content-Type", "application/jar")

	// Send the zip file to the client
	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Unable to send file", http.StatusInternalServerError)
		return
	}
}

func getOutdated(w http.ResponseWriter, r *http.Request) {
	isRemoteVersion := r.URL.Query().Has("version")
	remoteVersion := r.URL.Query().Get("version")
	fmt.Printf(getLocalVersion() + " vs " + remoteVersion)
	if isRemoteVersion && (remoteVersion != getLocalVersion()) {
		io.WriteString(w, "true")
	} else {
		io.WriteString(w, "false")
	}
}

func postCreatePropFile(w http.ResponseWriter, r *http.Request) {
	item := r.PostFormValue("item")
	if item == "" {
		item = "air"
	}
	displayName := r.PostFormValue("displayName")
	if displayName == "" {
		displayName = ""
	}
	texture := r.PostFormValue("texture")
	if texture == "" {
		texture = ""
	}
	fileName := r.PostFormValue("fileName")
	if fileName == "" {
		fileName = "error"
	}
	makePropFile(item, displayName, texture, fileName)
	io.WriteString(w, "OK!")
}

func postCreateElytraPropFile(w http.ResponseWriter, r *http.Request) {
	displayName := r.PostFormValue("displayName")
	if displayName == "" {
		displayName = ""
	}
	texture := r.PostFormValue("texture")
	if texture == "" {
		texture = ""
	}
	fileName := r.PostFormValue("fileName")
	if fileName == "" {
		fileName = "error"
	}
	makeElytraPropFile(displayName, texture, fileName)
	io.WriteString(w, "OK!")
}

func postCreateTexture(w http.ResponseWriter, r *http.Request) {
	url := r.PostFormValue("url")
	if url == "" {
		url = ""
	}
	fileName := r.PostFormValue("fileName")
	if fileName == "" {
		fileName = "error"
	}
	saveImage(url, fileName)
	io.WriteString(w, "OK!")
}

func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("api_key")
		keyHash := sha256.Sum256([]byte(key))
		expectedKeyHash := sha256.Sum256([]byte(apiKey))

		keyMatch := (subtle.ConstantTimeCompare(keyHash[:], expectedKeyHash[:]) == 1)

		fmt.Println(key)
		fmt.Println(apiKey)
		fmt.Println(keyMatch)

		if keyMatch {
			next.ServeHTTP(w, r)
			return
		}

		w.WriteHeader(http.StatusTeapot)
	})
}
