package apod

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// Apod response type from nasa endpoint
type Apod struct {
	Hdurl       string `json:"hdurl"`
	MediaType   string `json:"media_type"`
	Title       string `json:"title"`
	Explanation string `json:"explanation"`
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

	switch ext {
	case ".jpg":
		return a.appendTextToJpg(out, imageResponse.Body)
	default:
		_, err = io.Copy(out, imageResponse.Body)
		return err
	}
}

func (a *Apod) appendTextToJpg(w io.Writer, r io.Reader) error {
	fmt.Println("Start decoding image as jpeg")
	img, err := jpeg.Decode(r)
	if err != nil {
		return err
	}

	fmt.Println("Creating RGBA canvas big as the image")
	imgBound := img.Bounds()
	bitmap := image.NewRGBA(image.Rect(0, 0, imgBound.Dx(), imgBound.Dy()))

	fmt.Println("Painting image")
	draw.Draw(bitmap, bitmap.Bounds(), img, imgBound.Min, draw.Src)

	fmt.Printf("Adding text to canvas:\n%s\n", a.Explanation)
	startX := fixed.Int26_6(64*imgBound.Dx()) / 2
	startY := fixed.Int26_6(64*imgBound.Dy()) / 2
	d := &font.Drawer{
		Dst:  bitmap,
		Src:  image.NewUniform(color.White),
		Face: basicfont.Face7x13,
		Dot: fixed.Point26_6{
			X: startX,
			Y: startY,
		},
	}

	drawRows(d, 13, imgBound.Dy(), startX, startY, a.Explanation)

	fmt.Println("Encoding new image as jpeg")
	return jpeg.Encode(w, bitmap, &jpeg.Options{Quality: 95})
}

func drawRows(d *font.Drawer, fontHeight, imgDy int, startX, startY fixed.Int26_6, label string) {

	measure := d.MeasureString(label)
	countRows := int(measure / (fixed.Int26_6(64*imgDy) / 2))
	fmt.Printf("Text measure: %v\nRows needed: %v\n", measure, countRows)
	if countRows <= 1 {
		d.DrawString(label)
		return
	}

	letters := int(len(label) / countRows)
	for i := 0; i <= countRows; i++ {
		d.Dot.Y = startY + (fixed.Int26_6(i) * fixed.Int26_6(64*(fixed.Int26_6(fontHeight))))
		d.Dot.X = startX
		if i == countRows {
			d.DrawString(label[i*letters:])
		} else {
			d.DrawString(label[i*letters : (i+1)*letters])
		}
	}
}
