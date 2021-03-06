package main

import (
	"flag"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
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

type cursor struct {
	file int
	img  int
}

func (cur *cursor) Next(nbFiles int) {
	cur.file = (cur.file + 1) % nbFiles
	cur.img = 0
}

func (cur *cursor) Prev(nbFiles int) {
	cur.file = (cur.file - 1)
	if cur.file < 0 {
		cur.file = nbFiles + cur.file
	}
	cur.img = 0
}

// Current displayed file and image in file.
var cur = cursor{file: 0, img: 0}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: script FILE ...")
	}
	log.SetFlags(0)
	log.SetPrefix("[view-fits] ")

	const appID = "com.github.saimn.fitsview"
	application, err := gtk.ApplicationNew(appID, glib.APPLICATION_HANDLES_COMMAND_LINE)
	if err != nil {
		log.Fatal("Could not create application:", err)
	}

	application.Connect("command-line", func() int {
		flag.Parse()
		fmt.Printf("Args: %v\n", flag.Args())
		application.Activate()
		return 0
	})

	application.Connect("activate", func() {
		win := newWindow(application)

		// aNew := glib.SimpleActionNew("new", nil)
		// aNew.Connect("activate", func() {
		// 	newWindow(application).ShowAll()
		// })
		// application.AddAction(aNew)

		aQuit := glib.SimpleActionNew("quit", nil)
		aQuit.Connect("activate", func() {
			application.Quit()
		})
		application.AddAction(aQuit)

		win.ShowAll()
	})

	os.Exit(application.Run(os.Args))
}

func newWindow(application *gtk.Application) *gtk.ApplicationWindow {
	infos := processFiles()
	nbFiles := len(infos)
	if len(infos) == 0 {
		log.Fatal("No image among given FITS files.")
	}

	win, err := gtk.ApplicationWindowNew(application)
	if err != nil {
		log.Fatal("Unable to create window:", err)
	}
	win.SetTitle("FITSview")

	// Create a header bar
	header, err := gtk.HeaderBarNew()
	if err != nil {
		log.Fatal("Could not create header bar:", err)
	}
	header.SetShowCloseButton(true)
	header.SetTitle("GOTK3")

	// Create a new menu button
	mbtn, err := gtk.MenuButtonNew()
	if err != nil {
		log.Fatal("Could not create menu button:", err)
	}

	// Set up the menu model for the button
	menu := glib.MenuNew()
	if menu == nil {
		log.Fatal("Could not create menu (nil)")
	}
	// Actions with the prefix 'app' reference actions on the application
	// Actions with the prefix 'win' reference actions on the current window (specific to ApplicationWindow)
	// Other prefixes can be added to widgets via InsertActionGroup
	// menu.Append("New Window", "app.new")
	// menu.Append("Close Window", "win.close")
	menu.Append("Next file [right]", "custom.nextfile")
	menu.Append("Prev file [left]", "custom.prevfile")
	menu.Append("Quit", "app.quit")

	// Create the action "win.close"
	aClose := glib.SimpleActionNew("close", nil)
	aClose.Connect("activate", func() {
		win.Close()
	})
	win.AddAction(aClose)

	// Create and insert custom action group with prefix "custom"
	customActionGroup := glib.SimpleActionGroupNew()
	win.InsertActionGroup("custom", customActionGroup)

	// add the menu button to the header
	mbtn.SetMenuModel(&menu.MenuModel)
	header.PackStart(mbtn)
	win.SetTitlebar(header)

	vbox, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	win.Add(vbox)

	imageWidget, err := gtk.ImageNew()
	vbox.Add(imageWidget)
	// win.SetDefault(imageWidget)

	vbox.PackStart(footerBar(), false, false, 5)

	drawImage := func(i int) {
		log.Printf("file: %v\n", infos[i].Name)
		log.Printf("ext : %d/%d\n", cur.img+1, len(infos[i].Images))
		header.SetSubtitle(infos[i].Name)
		img := &infos[i].Images[cur.img]
		qmin, qmax := computeQuantiles(img, 0.01, 0.99)
		pixbuf, _ := pixBufFromImage(img.Image, qmin, qmax)
		imageWidget.SetFromPixbuf(pixbuf)
	}
	drawImage(cur.file)

	// Create an action in the custom action group
	aNextFile := glib.SimpleActionNew("nextfile", nil)
	aNextFile.Connect("activate", func() {
		cur.Next(nbFiles)
		drawImage(cur.file)
	})
	customActionGroup.AddAction(aNextFile)
	win.AddAction(aNextFile)

	aPrevFile := glib.SimpleActionNew("prevfile", nil)
	aPrevFile.Connect("activate", func() {
		cur.Prev(nbFiles)
		drawImage(cur.file)
	})
	customActionGroup.AddAction(aPrevFile)
	win.AddAction(aPrevFile)

	keyMap := map[uint]func(){
		gdk.KEY_q: func() {
			application.Quit()
		},
		gdk.KEY_Left: func() {
			cur.Prev(nbFiles)
			drawImage(cur.file)
		},
		gdk.KEY_Right: func() {
			cur.Next(nbFiles)
			drawImage(cur.file)
		},
		gdk.KEY_Up: func() {
			if len(infos[cur.file].Images) > 1 {
				cur.img = (cur.img + 1) % len(infos[cur.file].Images)
				drawImage(cur.file)
			}
		},
		gdk.KEY_Down: func() {
			if len(infos[cur.file].Images) > 1 {
				cur.img = (cur.img - 1)
				if cur.img < 0 {
					cur.img = len(infos[cur.file].Images) + cur.img
				}
				drawImage(cur.file)
			}
		},
	}

	win.Connect("key-press-event", func(win *gtk.ApplicationWindow, ev *gdk.Event) {
		keyEvent := &gdk.EventKey{ev}
		if move, found := keyMap[keyEvent.KeyVal()]; found {
			move()
			win.QueueDraw()
		}
	})

	// Set the default window size.
	img := &infos[cur.file].Images[cur.img]
	win.SetDefaultSize(img.Bounds().Dx(), img.Bounds().Dy())

	return win
}

func footerBar() *gtk.Box {
	hbox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)

	minLab, err := gtk.LabelNew("qmin")
	if err != nil {
		log.Fatal("Unable to create label:", err)
	}
	hbox.PackStart(minLab, false, false, 10)
	minBtn, err := gtk.SpinButtonNewWithRange(0, 100, 0.5)
	if err != nil {
		log.Fatal("Unable to create spin button:", err)
	}
	minBtn.SetValue(1)
	hbox.PackStart(minBtn, false, false, 10)

	maxLab, err := gtk.LabelNew("qmax")
	if err != nil {
		log.Fatal("Unable to create label:", err)
	}
	hbox.PackStart(maxLab, false, false, 10)
	maxBtn, err := gtk.SpinButtonNewWithRange(0, 100, 0.5)
	if err != nil {
		log.Fatal("Unable to create spin button:", err)
	}
	maxBtn.SetValue(99)
	hbox.PackStart(maxBtn, false, false, 10)

	minBtn.Connect("value-changed", func(sb *gtk.SpinButton) {
		fmt.Printf("val: %v\n", float64(sb.GetValue()))
	})

	maxBtn.Connect("value-changed", func(sb *gtk.SpinButton) {
		fmt.Printf("val: %v\n", float64(sb.GetValue()))
	})

	return hbox
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

func computeQuantiles(img image.Image, qmin, qmax float64) (float64, float64) {
	pixels, _ := getPixels(img)

	// Sort the values.
	inds := make([]int, len(pixels))
	floats.Argsort(pixels, inds)
	log.Printf("min: %v, max: %v\n", pixels[0], pixels[len(pixels)-1])

	// mean, std := stat.MeanStdDev(pixels, nil)
	// log.Printf("mean: %v, std: %v\n", mean, std)

	q1 := stat.Quantile(qmin, stat.Empirical, pixels, nil)
	q2 := stat.Quantile(qmax, stat.Empirical, pixels, nil)
	log.Printf("quantiles %.2f=%v, %.2f=%v\n", qmin, q1, qmax, q2)

	return q1, q2
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
