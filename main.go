package main

// Huge thanks the peson who reverse engineered this format:
//   http://www.shikadi.net/moddingwiki/Viacom_New_Media_Graphics_File_Format

import (
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/docopt/docopt-go"
)

const validVNMFSignatureBE = 0x564E4D1A
const validVNMFSignatureLE = 0x1A4D4E56
const TransparentPixel = 189

const (
	BitmapImage = iota
	SpriteImage
)

type Options struct {
	Info      bool   `docopt:"info"`
	InputFile string `docopt:"<inputfile>"`
	Extract   bool   `docopt:"extract"`
	OutputDir string `docopt:"<outputdir>"`
	Help      bool   `docopt:"--help"`
	Version   bool   `docopt:"--version"`
	ImageNum  int    `docopt:"--image"`
}

func main() {
	usage := `Viacom New Media Graphics File Exporter.

Usage:
  vnmf info [--image=<n>] <inputfile>
  vnmf extract [--image=<n>] <inputfile> <outputdir>
  vnmf -h | --help
  vnmf --version

Options:
  --image=<n>   The image number to view/extract
  -h --help     Show this screen.
  --version     Show version.`

	args, err := docopt.ParseArgs(usage, flag.Args(), "0.0.1")

	if err != nil {
		log.Fatalf("could not parse arguments: %s", err)
	}
	var options Options
	args.Bind(&options)

	if options.Info {
		info(options)
		return
	}

	if options.Extract {
		extract(options)
		return
	}
}

func info(options Options) {
	vnmf, err := OpenVNMFile(options.InputFile)
	if err != nil {
		log.Fatalf("could not open input file: %s", err)
	}

	if options.ImageNum > 0 {
		if options.ImageNum > int(len(vnmf.Images)) {
			log.Fatal("Invalid image index: Index out of range")
		}
		vnmi := vnmf.Images[options.ImageNum-1]
		t := "bitmap"
		if vnmi.Type == SpriteImage {
			t = "sprite"
		}
		fmt.Printf("Type: %s\nWidth: %dpx\nHeight: %dpx\n", t, vnmi.Width, vnmi.Height)
		return
	}

	fmt.Printf("Size: %v\nColors: %d\nImages: %d\n", vnmf.Size, vnmf.PaletteSize, vnmf.ImagesCount)
}

func extract(options Options) {
	vnmf, err := OpenVNMFile(options.InputFile)
	if err != nil {
		log.Fatal(err)
	}

	if options.ImageNum > 0 {
		if options.ImageNum > int(len(vnmf.Images)) {
			log.Fatal("Invalid image index: Index out of range")
		}
		vnmi := vnmf.Images[options.ImageNum-1]
		f, err := os.OpenFile(filepath.Join(options.OutputDir, fmt.Sprintf("img-%03d.png", options.ImageNum)), os.O_CREATE|os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatal(err)
		}

		if err := vnmi.Export(f); err != nil {
			log.Fatalf("could not export image: %s", err)
		}

		if err := f.Close(); err != nil {
			log.Fatalf("error writing to file: %s", err)
		}

	} else {
		for x, vnmi := range vnmf.Images {
			f, err := os.OpenFile(filepath.Join(options.OutputDir, fmt.Sprintf("img-%03d.png", x+1)), os.O_CREATE|os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				log.Fatal(err)
			}

			if err := vnmi.Export(f); err != nil {
				log.Fatalf("could not export image: %s", err)
			}

			if err := f.Close(); err != nil {
				log.Fatalf("error writing to file: %s", err)
			}
		}
	}
}

type VNMFile struct {
	*VNMFileHeader
	Palette *color.Palette
	Images  []*VNMImage
}

type VNMFileHeader struct {
	Signature         uint32 // file signature, should match 0x564E4D1A ("VNM\x1A") or 0x564E4D19
	Flags             uint32 // maybe file information flags? seems to be 0
	Size              uint32 // size of data
	PaletteOffset     uint32 // offset to palette data
	Unknown1Offset    uint32 // offset to unknown data (256 long array maybe?)
	Unknown2Offset    uint32 // offset to unknown data (256 long array maybe?)
	ImagesIndexOffset uint32 // offset to vnm image array
	PaletteStart      uint32 // 10 - first valid index in color palette
	PaletteSize       uint32 // 236 - Count of palette indexes stored in file at PaletteOffset
	ImagesCount       uint32 // 211 - Count of images stored in file at OffsetImages
}

type VNMImage struct {
	*VNMImageHeader
	Image   *image.Paletted
	Palette *color.Palette
	Number  int
}

func (vnmi *VNMImage) Export(w io.Writer) error {
	if err := png.Encode(w, vnmi.Image); err != nil {
		return fmt.Errorf("could not encode image as PNG: %s", err)
	}

	return nil
}

type VNMImageHeader struct {
	Offset uint32 // Offset in file to where raw image data starts
	Type   uint32 // 0 = Bitmap, 1 = Sprite
	Width  int32  // Width of the image
	Height int32  // Height of the imag<leader>Re
	XPos   int32  // x-position of image
	YPos   int32  // y-position of image
}

func OpenVNMFile(path string) (*VNMFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open file %s: %s", path, err)
	}
	defer f.Close()

	vnmf, err := FromVNMFile(f)
	if err != nil {
		return nil, fmt.Errorf("could not load vnm file: %s", err)
	}

	return vnmf, nil
}

func FromVNMFile(f *os.File) (*VNMFile, error) {
	vnmf := &VNMFile{VNMFileHeader: &VNMFileHeader{}}

	var err error

	err = populateVNMHeaderFromFile(vnmf, f)
	if err != nil {
		log.Fatal(err)
	}

	err = populatePaletteFromFile(vnmf, f)
	if err != nil {
		log.Fatal(err)
	}

	err = populateImagesFromFile(vnmf, f)
	if err != nil {
		log.Fatal(err)
	}

	return vnmf, nil
}

func populateVNMHeaderFromFile(vnmf *VNMFile, f *os.File) error {
	f.Seek(0, io.SeekStart)
	binary.Read(f, binary.LittleEndian, vnmf.VNMFileHeader)

	if vnmf.Signature != validVNMFSignatureLE {
		return errors.New("could not parse as vnmf format")
	}

	return nil
}

func populatePaletteFromFile(vnmf *VNMFile, f *os.File) error {
	f.Seek(int64(vnmf.PaletteOffset), io.SeekStart)

	palette := make(color.Palette, vnmf.PaletteSize+vnmf.PaletteStart)
	for i := 0; i < int(vnmf.PaletteStart); i++ {
		palette[i] = color.RGBA{}
	}
	for i := vnmf.PaletteStart; i < (vnmf.PaletteStart + vnmf.PaletteSize); i++ {
		rgb := make([]byte, 3)

		n, err := f.Read(rgb)
		if err != nil {
			return fmt.Errorf("could not read bytes: %s", err)
		}
		if n != 3 {
			return fmt.Errorf("something went wrong: wrong number of bytes read (got: %d, expected 3)", n)
		}

		palette[i] = color.RGBA{rgb[0] << 2, rgb[1] << 2, rgb[2] << 2, 0xff}
	}

	vnmf.Palette = &palette

	return nil
}

func populateImagesFromFile(vnmf *VNMFile, f *os.File) error {
	f.Seek(int64(vnmf.ImagesIndexOffset), io.SeekStart)
	imageCount := int(vnmf.ImagesCount)

	imageOffsetIndex := make([]uint32, imageCount)         //
	binary.Read(f, binary.LittleEndian, &imageOffsetIndex) // Byte offset for a given image

	images := make([]*VNMImage, imageCount)
	for i, offset := range imageOffsetIndex {
		vnmi := &VNMImage{VNMImageHeader: &VNMImageHeader{}}
		vnmi.Number = i + 1
		f.Seek(int64(offset), io.SeekStart)
		binary.Read(f, binary.LittleEndian, vnmi.VNMImageHeader)
		vnmi.Image = image.NewPaletted(image.Rect(0, 0, int(vnmi.Width), int(vnmi.Height)), *vnmf.Palette)
		if err := populateImageDataFromFile(vnmi, f); err != nil {
			return fmt.Errorf("error extracting bitmap data: %s", err)
		}
		images[i] = vnmi
	}
	vnmf.Images = images
	return nil
}

func populateImageDataFromFile(vnmi *VNMImage, f *os.File) error {
	var err error
	var pixels []uint8
	switch vnmi.Type {
	case BitmapImage:
		pixels, err = extractBitmapDataFromFile(vnmi, f)
	case SpriteImage:
		pixels, err = extractSpriteDataFromFile(vnmi, f)
	default:
		return fmt.Errorf("invalid image type identifier specified: %d", vnmi.Type)
	}
	if err != nil {
		return fmt.Errorf("error while extracting image data: %s", err)
	}

	populateImagePixels(vnmi, pixels)

	return nil
}

func populateImagePixels(vnmi *VNMImage, pixels []uint8) {
	h := int(vnmi.Height)
	w := int(vnmi.Width)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := (w * y) + x
			pixel := pixels[i]
			if pixel != TransparentPixel {
				vnmi.Image.SetColorIndex(x, y, pixel)
			}
		}
	}
}

func extractBitmapDataFromFile(vnmi *VNMImage, f *os.File) ([]uint8, error) {
	f.Seek(int64(vnmi.Offset), io.SeekStart)
	data := make([]uint8, vnmi.Width*vnmi.Height)

	w := int(vnmi.Width)
	h := int(vnmi.Height)
	i := 0
	for y := 0; y < h; y++ {
		row, err := ioutil.ReadAll(io.LimitReader(f, int64(w)))
		if err != nil {
			return nil, fmt.Errorf("error while reading bitmap row: %s", err)
		}
		for x := 0; x < w; x++ {
			data[i] = row[x]
			i++
		}
	}
	return data, nil
}

func extractSpriteDataFromFile(vnmi *VNMImage, f *os.File) ([]uint8, error) {
	rowOffsetIndex := make([]uint32, vnmi.Height) // byte offset index for a given row's pixel date
	f.Seek(int64(vnmi.Offset), io.SeekStart)
	binary.Read(f, binary.LittleEndian, &rowOffsetIndex)
	data := make([]uint8, vnmi.Width*vnmi.Height)
	escape := int32((0x100 - vnmi.Width)) // Used to signify we have a trapsparency run
	i := 0
	for y := 0; y < int(vnmi.Height); y++ {
		offset := rowOffsetIndex[y]
		f.Seek(int64(offset), io.SeekStart)
		row, err := ioutil.ReadAll(io.LimitReader(f, int64(vnmi.Width)))
		if err != nil {
			return nil, fmt.Errorf("error while reading sprite row: %s", err)
		}
		width := vnmi.Width
		j := width
		r := 0
		for j > 0 {
			var pixel int32
			// TODO: Ugh. I don't know why we run off the end for some images.
			if r < len(row) {
				pixel = int32(row[r])
			} else {
				pixel = TransparentPixel
				//fmt.Printf("Bad pixel (%d, %d) in image number %d\n", r, y, vnmi.Number)
			}
			r++
			if pixel >= escape {
				runLength := 0x100 - pixel
				r++ // skip next byte
				for runLength > 0 && j > 0 {
					runLength--
					j--
					data[i] = TransparentPixel
					i++
				}
			} else if j == width && pixel < escape {
				width++
			} else {
				data[i] = uint8(pixel)
				i++
				j--
			}
		}
	}

	return data, nil
}
