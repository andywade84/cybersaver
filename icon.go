package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"sync"
)

var (
	iconOnce  sync.Once
	iconCache []byte
)

func iconBytes() []byte {
	iconOnce.Do(func() {
		img := image.NewRGBA(image.Rect(0, 0, 16, 16))
		primary := color.RGBA{0, 255, 200, 255}
		dark := color.RGBA{15, 20, 32, 255}
		draw.Draw(img, img.Bounds(), &image.Uniform{dark}, image.Point{}, draw.Src)
		draw.Draw(img, image.Rect(1, 1, 15, 15), &image.Uniform{primary}, image.Point{}, draw.Src)
		draw.Draw(img, image.Rect(5, 5, 11, 11), &image.Uniform{dark}, image.Point{}, draw.Src)

		var pngBuf bytes.Buffer
		_ = png.Encode(&pngBuf, img)
		pngData := pngBuf.Bytes()

		var buf bytes.Buffer
		// ICONDIR
		binary.Write(&buf, binary.LittleEndian, uint16(0)) // reserved
		binary.Write(&buf, binary.LittleEndian, uint16(1)) // type = icon
		binary.Write(&buf, binary.LittleEndian, uint16(1)) // count
		// ICONDIRENTRY
		buf.WriteByte(16)                                             // width
		buf.WriteByte(16)                                             // height
		buf.WriteByte(0)                                              // color count
		buf.WriteByte(0)                                              // reserved
		binary.Write(&buf, binary.LittleEndian, uint16(1))            // planes
		binary.Write(&buf, binary.LittleEndian, uint16(32))           // bitcount
		binary.Write(&buf, binary.LittleEndian, uint32(len(pngData))) // bytes in resource
		binary.Write(&buf, binary.LittleEndian, uint32(6+16))         // offset
		buf.Write(pngData)

		iconCache = buf.Bytes()
	})
	return iconCache
}
