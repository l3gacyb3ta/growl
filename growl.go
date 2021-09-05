package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/docopt/docopt-go"
	"github.com/gabriel-vasile/mimetype"
	"github.com/jochasinga/requests"
)

var (
	baseURL = "https://pat.doggo.ninja/v1/"
)

// User is a struct for userdata
type User struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Admin           bool   `json:"admin"`
	Usage           int64  `json:"usage"`
	PreferredDomain string `json:"preferrredDomain"`
}

// pretty print user information
func (userInfo User) print() {
	fmt.Println("Username: " + userInfo.Name)
	fmt.Println("Usage: " + ByteCountSI(userInfo.Usage))

	if userInfo.Admin {
		fmt.Println("You are an admin")
	}
}

// A File is a struct for file information
type File struct {
	URL              string `json:"url"`
	ShortName        string `json:"shortName"`
	OriginalFileName string `json:"originalName"`
	Mime             string `json:"mimeType"`
	Size             int64  `json:"size"`
}

type uploadResponse struct {
	URL  string `json:"url"`
	Size uint64 `json:"size"`
}

func (file File) print() {
	fmt.Println("- Name: " + file.OriginalFileName)
	fmt.Println("\t URL:  " + file.URL)
	fmt.Println("\t Size: " + ByteCountSI(file.Size))

}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// ByteCountSI creates a human readable byte rep
func ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

// getUser returns the user information
func getUser() User {
	auth := func(r *requests.Request) {
		r.Header.Add("Authorization", "Bearer "+os.Getenv("DOGGO_TOKEN"))
	}

	url := baseURL + "me"
	res, _ := requests.Get(url, auth)

	var userData User

	json.Unmarshal(res.JSON(), &userData)
	return userData
}

// getFiles will create an array with Files that have the info in them
func getFiles() []File {
	auth := func(r *requests.Request) {
		r.Header.Add("Authorization", "Bearer "+os.Getenv("DOGGO_TOKEN"))
	}

	// Create the request
	url := baseURL + "files"
	res, err := requests.Get(url, auth)
	if err != nil {
		panic(err)
	}

	var files []File

	// Create the array of files
	json.Unmarshal(res.JSON(), &files)
	return files
}

// Print out the files
func printFiles(files []File) {
	for _, file := range files {
		// use the File struct print function
		file.print()
	}
}

// delete a file based on it's original file name
func delete(originalName string) {
	auth := func(r *requests.Request) {
		r.Header.Add("Authorization", "Bearer "+os.Getenv("DOGGO_TOKEN"))
	}

	files := getFiles()

	for _, file := range files {
		if file.OriginalFileName == originalName {
			shortName := file.ShortName

			url := baseURL + "file/" + shortName

			_, err := requests.Delete(url, auth)
			if err != nil {
				panic(err)
			}

			fmt.Println(originalName, "deleted!")
			return
		}
	}
	fmt.Println(originalName, "not found :(")
}

// deleteAll dies what ut says on the tin, it deletes all the files
func deleteAll() {
	files := getFiles()

	for _, file := range files {
		fmt.Println("Deleting", file.OriginalFileName, "....")
		delete(file.OriginalFileName)
	}

	fmt.Println("Mass delete finished")
}

// This uploadFilePOST handles some dark magic of file uploads
func uploadFilePOST(url string, filename string) (string, []byte) {

	client := &http.Client{}
	data, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	req, err := http.NewRequest("POST", url, data)
	if err != nil {
		log.Fatal(err)
	}

	//auth
	req.Header.Add("Authorization", "Bearer "+os.Getenv("DOGGO_TOKEN"))
	// Why kog
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return resp.Status, content
}

// parseMtype takes the mtype and does string sorcery to make it url friendly
func parseMtype(mtype string) string {
	return strings.Replace(strings.Split(mtype, ";")[0], "/", "%2F", 1)
}

// uploadFile uploads a file at the given path
func uploadFile(path string) {
	file := filepath.Base(path)

	mtype, err := mimetype.DetectFile(path)
	check(err)
	mimeType := parseMtype(mtype.String())

	_, resp := uploadFilePOST("https://pat.doggo.ninja/v1/upload?mimeType="+mimeType+"&originalName="+file, path)

	var respStruct uploadResponse

	json.Unmarshal(resp, &respStruct)

	fmt.Println("New URL:", respStruct.URL)
	fmt.Println("Size:   ", ByteCountSI(int64(respStruct.Size)))
}

func main() {
	usage := `growl: a tool for interacting with doggo.ninja
Usage:
	growl
	growl ls
	growl user
	growl upload [-d | --dir] <path>
	growl delete [--all] [<originalName>]
	growl -v
Options:
	<path>  Optional path argument.
	<originalName>  The original name of the file to be manipulated.`

	opts, _ := docopt.ParseArgs(usage, os.Args[1:], "1.0.0")
	path, _ := opts.String("<path>")
	originalName, _ := opts.String("<originalName>")

	if opts["upload"] == true {
		println("Uploading", path, "...")
		uploadFile(path)
	} else if opts["user"] == true {
		getUser().print()
	} else if opts["ls"] == true {
		printFiles(getFiles())
	} else if opts["delete"] == true {
		fmt.Println("Deleting", originalName)
		delete(originalName)
	}
}
