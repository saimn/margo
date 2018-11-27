package rot13

import "io"

type reader struct {
	r io.Reader
}

func (rr reader) Read(p []byte) (int, error) {
	n, err := rr.r.Read(p)
	for i := 0; i < n; i++ {
		p[i] = rot13(p[i])
	}
	return n, err
}

// NewReader plop
func NewReader(r io.Reader) io.Reader {
	return reader{r}
}

func rot13(b byte) byte {
	// for i := 0; i < 256; i++ {
	// 	fmt.Println(i, string(i))
	// }

	var a byte
	if (b >= byte('a')) && (b <= byte('z')) {
		a = 'a'
	} else if (b >= byte('A')) && (b <= byte('Z')) {
		a = 'A'
	} else {
		return b
	}
	return (b-a+13)%26 + a
}
