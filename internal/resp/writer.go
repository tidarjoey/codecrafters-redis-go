package resp

import (
	"fmt"
	"io"
)

func WriteBulk(w io.Writer, s string) error {
	_, err := fmt.Fprintf(w, "$%d\r\n%s\r\n", len(s), s)
	return err
}

func WriteNullBulk(w io.Writer) error {
	_, err := io.WriteString(w, "$-1\r\n")
	return err
}

func WriteSimple(w io.Writer, s string) error {
	_, err := fmt.Fprintf(w, "+%s\r\n", s)
	return err
}

func WriteError(w io.Writer, s string) error {
	_, err := fmt.Fprintf(w, "-%s\r\n", s)
	return err
}

func WriteInteger(w io.Writer, n int) error {
	_, err := fmt.Fprintf(w, ":%d\r\n", n)
	return err
}
