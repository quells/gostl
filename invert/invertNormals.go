package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"

	stl "github.com/quells/gostl"
)

func main() {
	var input string
	flag.StringVar(&input, "i", "", "Input STL filepath")
	var output string
	flag.StringVar(&output, "o", "", "Output STL filepath")
	flag.Parse()

	if input == "" || output == "" {
		flag.Usage()
		os.Exit(1)
	}

	m, err := stl.ParseStlFile(input)
	if err != nil {
		log.Fatal(err)
	}
	if m == nil {
		log.Fatal(fmt.Errorf("Model is nil"))
	}

	for i, t := range m.Triangles {
		if i > 0 {
			break
		}
		n0 := t.Normal
		n1 := cross(t.P0, t.P1)
		if dot(n0, n1) > 0 {
			n1 = cross(t.P1, t.P0)
		}
		t.Normal = normalize(n1)
		m.Triangles[i] = t
	}

	err = m.WriteToFile(output)
	if err != nil {
		log.Fatal(err)
	}
}

func dot(a, b [3]float32) float32 {
	return a[0]*b[0] + a[1]*b[1] + a[2]*b[2]
}

func cross(a, b [3]float32) [3]float32 {
	result := [3]float32{}
	result[0] = a[1]*b[2] - b[1]*a[2]
	result[1] = a[2]*b[0] - b[2]*a[0]
	result[2] = a[0]*b[1] - b[0]*a[1]
	return result
}

func normalize(a [3]float32) [3]float32 {
	l := float32(math.Sqrt(float64(dot(a, a))))
	il := 1.0 / l
	for i, c := range a {
		a[i] = c * il
	}
	return a
}
