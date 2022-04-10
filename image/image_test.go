package image

import (
	"bytes"
	"os"
	"strconv"
	"testing"
)

func TestNewImage(t *testing.T) {
	images, _ := NewImages("../counters/m")
	image, err := images.GetImage(3)
	if err != nil {
		t.Fatal(err)
	}
	for _, i := range []int{1, 2} {

		fname := strconv.Itoa(i)
		out, _ := os.Create(fname + ".png")
		_ = image.Draw(out, i*10, i*100)
		out.Close()
	}

}

func BenchmarkDraw(b *testing.B) {
	image, err := NewImage("../counters/m/counter1.gif")
	if err != nil {
		b.Fatal(err)
	}
	var buf bytes.Buffer
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err = image.Draw(&buf, 10, 212); err != nil {
			b.Fatal(err)
		}
	}
}
