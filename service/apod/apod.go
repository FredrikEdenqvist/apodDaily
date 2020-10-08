package apod

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Apod response type from nasa endpoint
type Apod struct {
	Hdurl     string `json:"hdurl"`
	MediaType string `json:"media_type"`
}

// GetImageApodMetaData :
func GetImageApodMetaData(uri string) (*Apod, error) {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode > 200 {
		fmt.Println(res.Status)
		return nil, fmt.Errorf("unsuccessfull request: %s", res.Status)
	}

	target := &Apod{}

	err = json.NewDecoder(res.Body).Decode(target)
	if err != nil {
		return nil, err
	}

	if target.MediaType != "image" {
		return nil, fmt.Errorf("unsupported media type: %s", target.MediaType)
	}

	return target, nil
}

// GetImageContent :
func (a *Apod) GetImageContent(path string) error {
	imageResponse, err := http.Get(a.Hdurl)
	if err != nil {
		return err
	}
	if imageResponse.StatusCode > 200 {
		log.Printf("unable to fetch %s %s\n", a.Hdurl, imageResponse.Status)
		return fmt.Errorf("unable to download: %s", imageResponse.Status)
	}
	defer imageResponse.Body.Close()

	s := time.Now().String()[0:19]
	s = strings.Replace(s, "-", "", 2)
	s = strings.Replace(s, " ", "", 1)
	s = strings.Replace(s, ":", "", 2)

	ext := filepath.Ext(a.Hdurl)

	// Create the file
	out, err := os.Create(path + s + ext)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, imageResponse.Body)
	return err
}
