package tga

import (
	"encoding/binary"
	"errors"
	"image"
	"image/draw"
	"io"
)

// Encode encodes an image into TARGA format.
func Encode(w io.Writer, m image.Image) (err error) {
	b := m.Bounds()
	mw, mh := b.Dx(), b.Dy()

	h := rawHeader{
		Width:  uint16(mw),
		Height: uint16(mh),
	}

	if int(h.Width) != mw || int(h.Height) != mh {
		return errors.New("uint16 width/height overflow")
	}

	h.Flags = flagOriginTop

	switch tm := m.(type) {
	case *image.Gray:
		h.ImageType = imageTypeMonoChrome
		err = encodeGray(w, tm, h)

	case *image.NRGBA:
		h.ImageType = imageTypeTrueColor
		err = encodeRGBA(w, tm, h, attrTypeNoAlpha)

	case *image.RGBA:
		h.ImageType = imageTypeTrueColor
		err = encodeRGBA(w, (*image.NRGBA)(tm), h, attrTypeNoAlpha)

	default:
		// convert to non-premultiplied alpha by default
		h.ImageType = imageTypeTrueColor
		newm := image.NewNRGBA(b)
		draw.Draw(newm, b, m, b.Min, draw.Src)
		err = encodeRGBA(w, newm, h, attrTypeNoAlpha)
	}

	return
}

func encodeGray(w io.Writer, m *image.Gray, h rawHeader) (err error) {
	h.BPP = 8 // 8-bit monochrome

	if err = binary.Write(w, binary.LittleEndian, &h); err != nil {
		return
	}

	offset := -(m.Rect.Min.Y*m.Stride + m.Rect.Min.X)
	max := offset + int(h.Height)*m.Stride

	for ; offset < max; offset += m.Stride {
		if _, err = w.Write(m.Pix[offset : offset+int(h.Width)]); err != nil {
			return
		}
	}

	// no extension area, only a footer
	err = binary.Write(w, binary.LittleEndian, newFooter())

	return
}

func encodeRGBA(w io.Writer, m *image.NRGBA, h rawHeader, attrType byte) (err error) {
	h.BPP = 24   // always save as 24-bit (faster this way)
	//h.Flags |= 8 // 8-bit alpha channel

	if err = binary.Write(w, binary.LittleEndian, &h); err != nil {
		return
	}

	lineSize := int(h.Width) * 4
	dstLineSize := int(h.Width) * 3
	offset := -m.Rect.Min.Y*m.Stride - m.Rect.Min.X*4
	max := offset + int(h.Height)*m.Stride
	b := make([]byte, dstLineSize)

	for ; offset < max; offset += m.Stride {
		var j int
		for i := 0; i < lineSize; i += 4 {
			// BGR
			b[j+0] = m.Pix[offset+i+2]
			b[j+1] = m.Pix[offset+i+1]
			b[j+2] = m.Pix[offset+i+0]
			j += 3
		}

		if _, err = w.Write(b); err != nil {
			return
		}
	}

	// add extension area and footer to define attribute type
	_, err = w.Write(newExtArea(attrType))
	footer := newFooter()
	footer.ExtAreaOffset = uint32(tgaRawHeaderSize + int(h.Height)*lineSize)
	err = binary.Write(w, binary.LittleEndian, footer)

	return
}
