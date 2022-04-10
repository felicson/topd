package image

//go:generate go-bindata -pkg $GOPACKAGE -o font2.go 5x8.bdf

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"io"
	"os"
	"path"
	"strconv"
	"sync"

	"github.com/zachomedia/go-bdf"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

//Image struct for render counter image
type Image struct {
	image image.Image //decoded image
	font  *bdf.Font
}

type ImageList map[uint]Image

//NewImage constructor
func NewImage(path string) (Image, error) {

	img, err := os.Open(path)
	if err != nil {
		return Image{}, err
	}
	defer img.Close()

	imgDecoded, err := gif.Decode(img)
	if err != nil {
		return Image{}, err
	}
	f, err := bdf.Parse(__5x8Bdf)
	if err != nil {
		return Image{}, err
	}
	return Image{image: imgDecoded, font: f}, nil
}

func (i Image) Draw(w io.Writer, hits, hosts int) error {

	const delim = "               "
	const lenDelim = 15
	var (
		hitsS  string
		hostsS string
	)

	if hits > 0 {
		hitsS = strconv.Itoa(hits)
	}
	if hosts > 0 {
		hostsS = strconv.Itoa(hosts)
	}

	col := color.RGBA{0, 0, 0, 255}
	point := fixed.P(6, 26)
	newImage := image.NewRGBA(i.image.Bounds())
	draw.Draw(newImage, i.image.Bounds(), i.image, image.ZP, draw.Src)
	d := font.Drawer{
		Dst:  newImage,
		Src:  image.NewUniform(col),
		Face: i.font.NewFace(),
		Dot:  point,
	}
	d.DrawString(hitsS + delim[len(hitsS):lenDelim-len(hostsS)] + hostsS)

	enc := &png.Encoder{
		CompressionLevel: png.NoCompression,
		//BufferPool:       bufPool,
	}
	if err := enc.Encode(w, newImage); err != nil {
		return err
	}
	return nil
}

var bufPool = sync.Pool{
	New: func() interface{} {
		// The Pool's New function should generally only return pointer
		// types, since a pointer can be put into the return interface
		// value without an allocation:
		return &png.EncoderBuffer{}
	},
}

func NewImages(imagePath string) (ImageList, error) {

	imageList := make(map[uint]Image, 4)

	countersDir, err := os.Open(imagePath)

	if err != nil {
		return nil, err
	}

	defer countersDir.Close()

	fileList, err := countersDir.Readdir(-1)

	if err != nil {
		return nil, err
	}

	for _, f := range fileList {

		if ext := path.Ext(f.Name()); ext == ".gif" {
			id, _ := strconv.Atoi(f.Name()[7:8])
			img, err := NewImage(path.Join(imagePath, f.Name()))
			if err != nil {
				return nil, err
			}
			imageList[uint(id)] = img
		}
	}
	return imageList, nil
}

//GetImage get image by id for drawing
func (images ImageList) GetImage(ID uint) (Image, error) {
	if img, ok := images[ID]; ok {
		return img, nil
	}
	return Image{}, errors.New("image not found")
}
