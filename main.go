package main

import (
	"fmt"
	"html/template"
	"image/color"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/google/uuid"
)

func FileSave(r *http.Request) string {
	// left shift 32 << 20 which results in 32*2^20 = 33554432
	// x << y, results in x*2^y
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		fmt.Println("error parsing")
		return ""
	}
	uuid := uuid.NewString()
	fmt.Println(uuid)
	// Retrieve the file from form data
	f, h, err := r.FormFile("file")
	if err != nil {
		fmt.Println("error retrieving file")

		return ""
	}
	defer f.Close()
	path := filepath.Join(".", "files")
	_ = os.MkdirAll(path, os.ModePerm)
	fullPath := path + "/" + uuid + filepath.Ext(h.Filename)
	file, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		fmt.Println("error ", err)

		return ""

	}
	defer file.Close()
	// Copy the file to the destination path
	_, err = io.Copy(file, f)
	if err != nil {
		return ""
	}
	return fullPath
}

func main() {

	fs := http.FileServer(http.Dir("files/"))
	http.Handle("/files/", http.StripPrefix("/files/", fs))

	tmpl, err := template.ParseFiles("static/index.html")
	fmt.Println(err)

	// Routes consist of a path and a handler function.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			tmpl.Execute(w, nil)
			return
		}
		filePath := FileSave(r)
		fmt.Println(r.FormValue("red"))
		red, err := strconv.Atoi(r.FormValue("red"))
		green, err := strconv.Atoi(r.FormValue("green"))
		blue, err := strconv.Atoi(r.FormValue("blue"))
		subject := Requestpic{filePath, r.FormValue("text"), uint8(red), uint8(green), uint8(blue)}
		Imgreturned, err := TextOnImg(subject)
		if err != nil {
			fmt.Println("error")
			tmpl.Execute(w, err)
		}
		fmt.Println(Imgreturned)
		tmpl.Execute(w, Imgreturned)
	})

	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(":8000", nil))

}

type Requestpic struct {
	BgImgPath  string
	Textinput  string
	TextColorR uint8
	TextColorG uint8
	TextColorB uint8
}

func pxTopt(pt float64) (px float64) {
	first := pt / 72.0
	second := first * 96.0
	return second
}

func TextOnImg(request Requestpic) (string, error) {
	bgImage, err := gg.LoadImage(request.BgImgPath)
	if err != nil {
		fmt.Println(err)
	}
	imgWidth := bgImage.Bounds().Dx()
	imgHeight := bgImage.Bounds().Dy()

	dc := gg.NewContext(imgWidth, imgHeight)
	dc.DrawImage(bgImage, 0, 0)
	//rect := drawrect(request)
	//dc.DrawImage(rect, imgWidth, imgHeight)

	fontsize := pxTopt(float64(imgHeight) / 10.0)

	if err := dc.LoadFontFace("Garamond.ttf", fontsize); err != nil {
		fmt.Println(err)
	}
	x := float64(imgWidth / 2)
	y := float64((imgHeight / 2) - int(fontsize)/2)

	ac := gg.NewContext(imgHeight, imgWidth)
	ac.DrawRectangle(0, 0, float64(imgHeight), float64(imgWidth)*1.2)
	grad := gg.NewLinearGradient(fontsize, 0, float64(imgHeight), 0)
	grad.AddColorStop(0, color.RGBA{0, 0, 0, 0})
	grad.AddColorStop(0.25, color.RGBA{0, 0, 0, 50})
	grad.AddColorStop(0.35, color.RGBA{0, 0, 0, 150})
	grad.AddColorStop(0.45, color.RGBA{0, 0, 0, 255})
	grad.AddColorStop(0.5, color.RGBA{0, 0, 0, 255})
	grad.AddColorStop(0.55, color.RGBA{0, 0, 0, 255})
	grad.AddColorStop(0.65, color.RGBA{0, 0, 0, 150})
	grad.AddColorStop(0.75, color.RGBA{0, 0, 0, 50})

	grad.AddColorStop(1, color.RGBA{0, 0, 0, 0})
	ac.SetFillStyle(grad)
	ac.Fill()
	//bc := DropShadow(ac.Image(), 50.0)
	ac.SavePNG("temp.png")
	cc := imaging.Rotate90(ac.Image())
	dc.DrawImageAnchored(cc, imgWidth/2, imgHeight/2, 0.5, 0.5)

	maxWidth := float64(imgWidth) - 60.0
	dc.SetColor(color.RGBA{request.TextColorR, request.TextColorG, request.TextColorB, 255})
	dc.DrawStringWrapped(request.Textinput, x, y*0.95, 0.5, 0.5, maxWidth, 1.5, gg.AlignCenter)

	if err := dc.LoadFontFace("Garamond.ttf", fontsize*1.05); err != nil {
		fmt.Println(err)
	}
	dc.SetColor(color.RGBA{request.TextColorR, request.TextColorG, request.TextColorB, 150})
	dc.DrawStringWrapped(request.Textinput, x, y*0.95, 0.5, 0.5, maxWidth, 1.5, gg.AlignCenter)

	dc.SavePNG(request.BgImgPath)
	return request.BgImgPath, nil
}
