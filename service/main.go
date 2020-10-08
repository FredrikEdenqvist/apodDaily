package main

import (
	"log"
	"net/url"
	"os"

	"github.com/FredrikEdenqvist/apodDaily/apod"
)

func main() {
	storagelocation := os.Getenv("APOD_LOCAL_STORE_LOCATION")
	if storagelocation == "" {
		log.Fatalln("no storage location defined")
		return
	}

	envAPIKEY := os.Getenv("APOD_API_KEY")
	params := url.Values{}

	if envAPIKEY != "" {
		params.Add("api_key", envAPIKEY)
	} else {
		params.Add("api_key", "DEMO_KEY")
	}

	uri := "https://api.nasa.gov/planetary/apod?" + params.Encode()

	apodData, err := apod.GetImageApodMetaData(uri)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = apodData.GetImageContent(storagelocation)
	if err != nil {
		log.Fatal(err)
		return
	}
}
