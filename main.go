package main

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/gorilla/mux"
)

func routeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "method is not supported", http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, "static/index.html")
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	file, handler, err := r.FormFile("file")
	fileName := r.FormValue("file_name")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	f, err := os.OpenFile(handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, _ = io.WriteString(w, "File "+fileName+" Uploaded successfully")
	_, _ = io.Copy(f, file)

	if err := r.ParseMultipartForm(8192); err != nil {
		fmt.Fprintf(w, "ParseMultipartForm() err: %v", err)
		return
	}
	fmt.Fprintf(w, "POST request successful")

	//text := r.FormValue("text")

}

func main() {
	text := "this is a test"
	testimage := Request{"img.png", "Garamond.ttf", text, 255, 0, 0}
	TextOnImg(testimage)
	r := mux.NewRouter()

	// Routes consist of a path and a handler function.
	r.HandleFunc("/", routeHandler).
		Methods("GET")
	r.HandleFunc("/submit", formHandler).
		Methods("POST")

	// Bind to a port and pass our router in
	log.Fatal(http.ListenAndServe(":8000", r))

}

type Request struct {
	BgImgPath  string
	FontPath   string
	Text       string
	TextColorR uint8
	TextColorG uint8
	TextColorB uint8
}

func pxTopt(pt float64) (px float64) {
	first := pt / 72.0
	second := first * 96.0
	return second
}

func TextOnImg(request Request) (image.Image, error) {
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

	if err := dc.LoadFontFace(request.FontPath, fontsize); err != nil {
		fmt.Println(err)
	}
	x := float64(imgWidth / 2)
	y := float64((imgHeight / 2) - int(fontsize)/2)

	ac := gg.NewContext(imgHeight, imgWidth)
	ac.DrawRectangle(0, 0, float64(imgHeight), float64(imgWidth))
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
	dc.DrawStringWrapped(request.Text, x, y, 0.5, 0.5, maxWidth, 1.5, gg.AlignCenter)

	if err := dc.LoadFontFace(request.FontPath, fontsize*1.05); err != nil {
		fmt.Println(err)
	}
	dc.SetColor(color.RGBA{request.TextColorR, request.TextColorG, request.TextColorB, 150})
	dc.DrawStringWrapped(request.Text, x, y, 0.5, 0.5, maxWidth, 1.5, gg.AlignCenter)

	dc.SavePNG("./out.png")
	return dc.Image(), nil
}
