package main

import (
	"flag"
	"image"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/saimn/fitsio"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/stat"
)

const maxUint32 = 65535

type fileInfo struct {
	Name   string
	Images []imageInfo
}

type imageInfo struct {
	image.Image
	scale int // image scale in percents (default: 100%)
	orig  image.Point
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: script FILE ...")
	}
	flag.Parse()
	log.SetFlags(0)
	log.SetPrefix("[view-fits] ")

	infos := processFiles()
	nbFiles := len(infos)
	if len(infos) == 0 {
		log.Fatal("No image among given FITS files.")
	}

	type cursor struct {
		file int
		img  int
	}

	// Current displayed file and image in file.
	cur := cursor{file: 0, img: 0}

	// Initialize GTK without parsing any command line arguments.
	gtk.Init(nil)

	// Create a new toplevel window, set its title, and connect it to the
	// "destroy" signal to exit the GTK main loop when it is destroyed.
	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	// Create a new label widget to show in the window.
	// l, err := gtk.LabelNew("Hello, gotk3!")
	// if err != nil {
	// 	log.Fatal("Unable to create label:", err)
	// }

	imageWidget, err := gtk.ImageNew()
	win.Add(imageWidget)

	drawImage := func(i int) {
		log.Printf("file: %v\n", infos[i].Name)
		log.Printf("ext : %d/%d\n", cur.img+1, len(infos[i].Images))
		win.SetTitle(infos[i].Name)
		img := &infos[i].Images[cur.img]
		pixels, _ := getPixels(img)

		// Sort the values.
		inds := make([]int, len(pixels))
		floats.Argsort(pixels, inds)
		log.Printf("min: %v, max: %v\n", pixels[0], pixels[len(pixels)-1])

		mean, std := stat.MeanStdDev(pixels, nil)
		log.Printf("mean: %v, std: %v\n", mean, std)

		quant1 := stat.Quantile(0.01, stat.Empirical, pixels, nil)
		quant99 := stat.Quantile(0.99, stat.Empirical, pixels, nil)
		log.Printf("quant1: %v, quant99: %v\n", quant1, quant99)

		pixbuf, _ := pixBufFromImage(img.Image, quant1, quant99)
		imageWidget.SetFromPixbuf(pixbuf)
	}
	drawImage(cur.file)

	keyMap := map[uint]func(){
		gdk.KEY_q: func() {
			gtk.MainQuit()
		},
		gdk.KEY_Left: func() {
			cur.file = (cur.file - 1)
			if cur.file < 0 {
				cur.file = nbFiles + cur.file
			}
			cur.img = 0
			drawImage(cur.file)
		},
		gdk.KEY_Right: func() {
			cur.file = (cur.file + 1) % nbFiles
			cur.img = 0
			drawImage(cur.file)
		},
		gdk.KEY_Up: func() {
			cur.img = (cur.img + 1) % len(infos[cur.file].Images)
			drawImage(cur.file)
		},
		gdk.KEY_Down: func() {
			cur.img = (cur.img - 1)
			if cur.img < 0 {
				cur.img = len(infos[cur.file].Images) + cur.img
			}
			drawImage(cur.file)
		},
	}

	win.Connect("key-press-event", func(win *gtk.Window, ev *gdk.Event) {
		keyEvent := &gdk.EventKey{ev}
		if move, found := keyMap[keyEvent.KeyVal()]; found {
			move()
			win.QueueDraw()
		}
	})

	// Set the default window size.
	img := &infos[cur.file].Images[cur.img]
	win.SetDefaultSize(img.Bounds().Dx(), img.Bounds().Dy())

	// Recursively show all widgets contained in this window.
	win.ShowAll()

	// Begin executing the GTK main loop.  This blocks until
	// gtk.MainQuit() is run.
	gtk.Main()
}

func processFiles() []fileInfo {
	infos := make([]fileInfo, 0, len(flag.Args()))
	// Parsing input files.
	for _, fname := range flag.Args() {

		finfo := fileInfo{Name: fname}

		r, err := openStream(fname)
		if err != nil {
			log.Fatalf("Can not open the input file: %v", err)
		}
		defer r.Close()

		// Opening the FITS file.
		f, err := fitsio.Open(r)
		if err != nil {
			log.Fatalf("Can not open the FITS input file: %v", err)
		}
		defer f.Close()

		// Getting the file HDUs.
		hdus := f.HDUs()
		for _, hdu := range hdus {
			// Getting the header informations.
			header := hdu.Header()
			axes := header.Axes()

			// Discarding HDU with no axes.
			if len(axes) != 0 {
				if hdu, ok := hdu.(fitsio.Image); ok {
					img := hdu.Image()
					if img != nil {
						finfo.Images = append(finfo.Images, imageInfo{
							Image: img,
							scale: 100,
							orig:  image.Point{},
						})
					}
				}
			}
		}

		if len(finfo.Images) > 0 {
			infos = append(infos, finfo)
		}
	}

	return infos
}

func openStream(name string) (io.ReadCloser, error) {
	switch {
	case strings.HasPrefix(name, "http://") || strings.HasPrefix(name, "https://"):
		resp, err := http.Get(name)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		f, err := ioutil.TempFile("", "view-fits-")
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(f, resp.Body)
		if err != nil {
			f.Close()
			return nil, err
		}

		// make sure we have at least a full FITS block
		f.Seek(0, 2880)
		f.Seek(0, 0)

		return f, nil

	case strings.HasPrefix(name, "file://"):
		name = name[len("file://"):]
		return os.Open(name)
	default:
		return os.Open(name)
	}
}

func pixBufFromImage(picture image.Image, vmin, vmax float64) (*gdk.Pixbuf, error) {
	width := picture.Bounds().Max.X
	height := picture.Bounds().Max.Y

	pixbuf, err := gdk.PixbufNew(gdk.COLORSPACE_RGB, true, 8, width, height)
	if nil != err {
		return nil, err
	}
	pixelSlice := pixbuf.GetPixels()

	const bytesPerPixel = 4
	i := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			colour := picture.At(x, y)
			r, _, _, _ := colour.RGBA()

			// scale
			var val uint32
			if r < uint32(vmin) {
				val = 0
			} else if r > uint32(vmax) {
				val = maxUint32
			} else {
				val = uint32((float64(r) - vmin) / (vmax - vmin) * maxUint32)
			}
			bval := uint32ToByte(val)
			pixelSlice[i] = bval   // r
			pixelSlice[i+1] = bval // g
			pixelSlice[i+2] = bval // b
			pixelSlice[i+3] = 255

			i += bytesPerPixel
		}
	}

	return pixbuf, nil
}

func uint32ToByte(value uint32) byte {
	const ratio = float64(256) / float64(65536)
	byteValue := ratio * float64(value)
	if byteValue > 255 {
		return byte(255)
	}
	return byte(byteValue)
}

// Get the bi-dimensional pixel array
func getPixels(img image.Image) ([]float64, error) {
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	var pixels []float64
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, _, _, _ := img.At(x, y).RGBA()
			if r != 0 {
				pixels = append(pixels, float64(r))
			}
		}
	}

	return pixels, nil
}

// img.At(x, y).RGBA() returns four uint32 values; we want a Pixel
// func rgbaToPixel(r uint32, g uint32, b uint32, a uint32) Pixel {
// 	return Pixel{int(r / 257), int(g / 257), int(b / 257), int(a / 257)}
// }
