package frame

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
)

// NativeFrame represents a native image frame
type NativeFrame struct {
	Rows            int
	Cols            int
	SamplesPerPixel int
	BitsPerSample   int
	// Data is a slice of pixels, where each pixel can have multiple values
	Data []byte
}

func (n *NativeFrame) IsEncapsulated() bool { return false }

func (n *NativeFrame) GetNativeFrame() (*NativeFrame, error) {
	return n, nil
}

func (n *NativeFrame) GetEncapsulatedFrame() (*EncapsulatedFrame, error) {
	return nil, ErrorFrameTypeNotPresent
}

func (n *NativeFrame) GetPixel(x, y int) (samples []uint32) {
	for i := 0; i < n.SamplesPerPixel; i++ {
		switch n.BitsPerSample {
		case 8:
			samples = append(samples, uint32(n.Data[(y*n.Cols+x)*n.SamplesPerPixel+i]))
		case 16:
			v := binary.LittleEndian.Uint16(n.Data[(y*n.Cols+x)*n.SamplesPerPixel*2+i*2:])
			samples = append(samples, uint32(v))
		case 32:
			v := binary.LittleEndian.Uint32(n.Data[(y*n.Cols+x)*n.SamplesPerPixel*4+i*4:])
			samples = append(samples, v)
		}
	}

	return
}

// GetImage returns an image.Image representation the frame, using default
// processing. This default processing is basic at the moment, and does not
// autoscale pixel values or use window width or level info.
func (n *NativeFrame) GetImage() (image.Image, error) {
	if n.SamplesPerPixel == 1 {
		img := image.NewGray16(image.Rect(0, 0, n.Cols, n.Rows))
		for i := 0; i < n.Cols*n.Rows; i++ {
			x, y := i%n.Cols, i/n.Cols
			img.SetGray16(x, y, color.Gray16{Y: uint16(n.GetPixel(x, y)[0])}) // for now, assume we're not overflowing uint16, assume gray image
		}
		return img, nil
	}
	return jpeg.Decode(bytes.NewReader(n.Data))
}
