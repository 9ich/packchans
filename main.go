package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"packchans/tga" //"github.com/ftrvxmtrx/tga"

	"github.com/nfnt/resize"
)

var outFile string

func init() {
	flag.StringVar(&outFile, "o", "packed.tga", "output filename")
}

func main() {
	flag.Parse()
	if flag.NArg() != 3 {
		fmt.Printf("usage: packchans -o out.tga red.tga green.tga blue.tga\n")
		flag.PrintDefaults()
		return
	}

	var imgs [3]image.Image
	for i := 0; i < len(imgs); i++ {
		var err error
		imgs[i], _, err = readImg(flag.Arg(i))
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	packed, err := pack(imgs)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = writeImg(outFile, packed)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func readImg(fname string) (image.Image, string, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()

	return image.Decode(bufio.NewReader(f))
}

func writeImg(fname string, img image.Image) error {
	f, err := os.Create(fname)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	return tga.Encode(f, img)
}

func pack(imgs [3]image.Image) (image.Image, error) {
	max := imgs[0]
	var maxArea int
	for i := 0; i < len(imgs); i++ {
		area := imgs[i].Bounds().Dx() * imgs[i].Bounds().Dy()
		if max == nil || area > maxArea {
			max = imgs[i]
			maxArea = max.Bounds().Dx() * max.Bounds().Dy()
		}
	}
	for i, img := range imgs {
		r := img.Bounds()
		mr := max.Bounds()
		if r.Dx() != mr.Dx() || r.Dy() != mr.Dy() {
			fmt.Printf("resize %d*%d -> %d*%d\n", r.Dx(), r.Dy(), mr.Dx(), mr.Dy())
			imgs[i] = resize.Resize(uint(mr.Dx()), uint(mr.Dy()), img, resize.Bilinear)
		}
	}

	out := image.NewNRGBA(max.Bounds())
	for y := 0; y < out.Bounds().Dy(); y++ {
		for x := 0; x < out.Bounds().Dx(); x++ {
			c := color.NRGBA{0, 0, 0, 0xFF}
			m := color.NRGBAModel
			c.R = m.Convert(imgs[0].At(x, y)).(color.NRGBA).R
			c.G = m.Convert(imgs[1].At(x, y)).(color.NRGBA).G
			c.B = m.Convert(imgs[2].At(x, y)).(color.NRGBA).B
			out.Set(x, y, c)
		}
	}
	return out, nil
}
