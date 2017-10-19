// Package gostl deals with binary STL files
package gostl

import (
	"fmt"
	"log"
	"math"
	"os"
)

const stlHeaderSize = 80   // bytes
const stlCountSize = 4     // bytes
const stlFloatSize = 4     // bytes
const stlTriangleSize = 50 // bytes

// Triangle is the smallest polygon
type Triangle struct {
	Normal, P0, P1, P2 [3]float32
}

// Model is a collection of triangles
type Model struct {
	filepath  string
	Triangles []Triangle
}

// BoundingBox is the axis-aligned box that contains a Model
func (m *Model) BoundingBox() (min, max [3]float32) {
	min = m.Triangles[0].P0
	max = m.Triangles[0].P0
	for _, t := range m.Triangles {
		min[0] = minimum(min[0], t.P0[0], t.P1[0], t.P2[0])
		min[1] = minimum(min[1], t.P0[1], t.P1[1], t.P2[1])
		min[2] = minimum(min[2], t.P0[2], t.P1[2], t.P2[2])
		max[0] = maximum(max[0], t.P0[0], t.P1[0], t.P2[0])
		max[1] = maximum(max[1], t.P0[1], t.P1[1], t.P2[1])
		max[2] = maximum(max[2], t.P0[2], t.P1[2], t.P2[2])
	}
	return
}

// WriteToFile writes a Model to a binary STL file
func (m *Model) WriteToFile(filepath string) error {
	writer, err := os.Create(filepath)
	defer writer.Close()
	if err != nil {
		return err
	}

	var buffer []byte

	// Write header
	buffer = make([]byte, stlHeaderSize)
	header := []byte("STL")
	for i, b := range header {
		buffer[i] = b
	}
	n, err := writer.Write(buffer)
	if err != nil {
		return err
	}
	if n < stlHeaderSize {
		return fmt.Errorf("could not write stl file header, unexpected header length")
	}

	// Write Triangle count
	n, err = writer.Write(bytesFrom(uint32(len(m.Triangles))))
	if err != nil {
		return err
	}
	if n < stlCountSize {
		return fmt.Errorf("could not write stl file header, unexpected Triangle count")
	}

	// Write Triangles
	buffer = make([]byte, stlTriangleSize)
	for _, t := range m.Triangles {
		for i := 0; i < 3; i++ {
			ui := math.Float32bits(t.Normal[i])
			bs := bytesFrom(ui)
			for j, b := range bs {
				buffer[j+4*i] = b
			}
		}
		for i := 0; i < 3; i++ {
			ui := math.Float32bits(t.P0[i])
			bs := bytesFrom(ui)
			for j, b := range bs {
				buffer[j+4*i+12] = b
			}
		}
		for i := 0; i < 3; i++ {
			ui := math.Float32bits(t.P1[i])
			bs := bytesFrom(ui)
			for j, b := range bs {
				buffer[j+4*i+24] = b
			}
		}
		for i := 0; i < 3; i++ {
			ui := math.Float32bits(t.P2[i])
			bs := bytesFrom(ui)
			for j, b := range bs {
				buffer[j+4*i+36] = b
			}
		}
		n, err = writer.Write(buffer)
		if err != nil {
			return err
		}
		if n < stlTriangleSize {
			return fmt.Errorf("could not write Triangle")
		}
	}

	return nil
}

// ParseStlFile parses a binary STL file into a Model
func ParseStlFile(filepath string) (*Model, error) {
	reader, err := os.Open(filepath)
	defer reader.Close()
	if err != nil {
		return nil, err
	}

	var buffer []byte

	// Ignore header
	buffer = make([]byte, stlHeaderSize)
	n, err := reader.Read(buffer)
	if err != nil {
		return nil, err
	}
	if n < stlHeaderSize {
		return nil, fmt.Errorf("could not read stl file header, unexpected header length")
	}

	// Read Triangle count
	buffer = make([]byte, stlCountSize)
	n, err = reader.Read(buffer)
	if err != nil {
		return nil, err
	}
	if n < stlCountSize {
		return nil, fmt.Errorf("could not read stl file header, unexpected count length")
	}
	numTriangles := uint32From(buffer)

	// Read Triangles
	Triangles := make([]Triangle, numTriangles)
	var i uint32
	for ; i < numTriangles; i++ {
		buffer = make([]byte, stlTriangleSize)
		n, err = reader.Read(buffer)
		if err != nil {
			return nil, err
		}
		if n < stlTriangleSize {
			return nil, fmt.Errorf("could not read stl file, unexpected Triangle byte count")
		}
		Triangles[i] = triangleFrom(buffer)
	}

	m := Model{filepath, Triangles}
	return &m, nil
}

func uint32From(bytes []byte) (x uint32) {
	if len(bytes) != 4 {
		log.Fatalf("expected 4 bytes, got %d: %q", len(bytes), bytes)
	}
	for i := 0; i < 4; i++ {
		x |= uint32(bytes[i]) << uint(8*i)
	}
	return
}

func bytesFrom(u uint32) []byte {
	bytes := make([]byte, 4)
	for i := 0; i < 4; i++ {
		bytes[i] = byte((u >> uint(8*i)) & 0xff)
	}
	return bytes
}

func triangleFrom(bytes []byte) Triangle {
	if len(bytes) != stlTriangleSize {
		log.Fatalf("expected %d bytes, got %d: %q", stlTriangleSize, len(bytes), bytes)
	}
	ps := [4][3]float32{}
	for j := 0; j < 4; j++ {
		for i := 0; i < 3; i++ {
			offset := i*stlFloatSize + j*3*stlFloatSize
			ui := uint32From(bytes[offset : offset+4])
			ps[j][i] = math.Float32frombits(ui)
		}
	}
	return Triangle{ps[0], ps[1], ps[2], ps[3]}
}

func minimum(xs ...float32) float32 {
	m := xs[0]
	for i := 1; i < len(xs); i++ {
		if xs[i] < m {
			m = xs[i]
		}
	}
	return m
}

func maximum(xs ...float32) float32 {
	M := xs[0]
	for i := 1; i < len(xs); i++ {
		if xs[i] > M {
			M = xs[i]
		}
	}
	return M
}
