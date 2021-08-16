package main

import (
	"encoding/json"
	"fmt"
	"os"
	
	"github.com/jochasinga/requests"
	//"os"
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

// A file
type File struct {
	URL string `json:"url"`
	ShortName string `json:"shortName"`
	OriginalFileName string `json:"originalName"`
	Mime string `json:"mimeType"`
	Size int64 `json:"size"`
}

func (file File) print() {
	fmt.Println("- Name: " + file.OriginalFileName)
	fmt.Println("\t URL:  " + file.URL)
	fmt.Println("\t Size: " + ByteCountSI(file.Size))

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
	for _, file := range(files) {
		file.print()
	}
}

func main() {
	// user := getUser()
	// printUser(user)

	files := getFiles()
	printFiles(files)
}
