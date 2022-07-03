package mbox

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/mail"
	"os"
)

// MessageListener gets the From_ line, including carriage return and line feed,
// plus the message in parsed and unparsed form. Implementations must not retain
// raw nor msg.Body.
type MessageListener func(fromLine string, raw []byte, msg *mail.Message)

// ErrNotMbox signals file rejection.
var ErrNotMbox = errors.New("not an mbox")

// ReadFile calls the listener for each entry read from file. The return is
// io.EOF if, and only if the file has no content.
func ReadFile(file string, onMessage MessageListener) error {
	// input stream
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	r := bufio.NewReader(f)

	// first From_ line
	line, err := r.ReadSlice('\n')
	switch {
	case err == nil:
		if !IsFromLine(line) {
			return fmt.Errorf("%s:1: %w", file, ErrNotMbox)
		}

	case errors.Is(err, bufio.ErrBufferFull):
		if string(line[:5]) != "From " {
			return fmt.Errorf("%s:1: %w", file, ErrNotMbox)
		}
		return fmt.Errorf("%s:1: From_ line exceeds %d bytes: %w", file, r.Size(), ErrNotMbox)

	case errors.Is(err, io.EOF):
		switch {
		case len(line) == 0:
			return io.EOF
		case line[0] != 'F',
			len(line) > 1 && line[1] != 'r',
			len(line) > 2 && line[2] != 'o',
			len(line) > 3 && line[3] != 'm',
			len(line) > 4 && line[4] != ' ':
			return fmt.Errorf("%s: %w", file, ErrNotMbox)
		default:
			return fmt.Errorf("%s:1: From_ line got %w", file, io.ErrUnexpectedEOF)
		}

	default:
		return err
	}

	fromLine := string(line)
	fromLineN := 1
	var buf bytes.Buffer

	for lineN := 2; ; lineN++ {
		line, err := r.ReadSlice('\n')
		switch {
		case err == nil:
			if !IsFromLine(line) {
				buf.Write(line)
				continue // hot-path
			}

			// call listener
			raw := buf.Bytes()
			msg, err := mail.ReadMessage(&buf)
			if err != nil {
				return fmt.Errorf("%s:%d–%d: %w", file, fromLineN, lineN-1, err)
			}
			onMessage(fromLine, raw, msg)

			// next
			fromLine = string(line)
			buf.Reset()

		case errors.Is(err, io.EOF):
			if len(line) != 0 {
				return fmt.Errorf("%s:%d: %w", file, lineN, io.ErrUnexpectedEOF)
			}

			// call listener (last time)
			raw := buf.Bytes()
			msg, err := mail.ReadMessage(&buf)
			if err != nil {
				return fmt.Errorf("%s:%d–%d: %w", file, fromLineN, lineN-1, err)
			}
			onMessage(fromLine, raw, msg)

			return nil // done

		case errors.Is(err, bufio.ErrBufferFull):
			buf.Write(line)
			err = copyLine(&buf, r)
			if err != nil {
				return fmt.Errorf("%s:%d: %w", file, lineN, err)
			}

		default:
			return fmt.Errorf("%s:%d: %w", file, lineN, err)
		}
	}
}

func copyLine(buf *bytes.Buffer, r *bufio.Reader) error {
	for {
		line, err := r.ReadSlice('\n')
		buf.Write(line)
		switch {
		case err == nil:
			return nil
		case errors.Is(err, bufio.ErrBufferFull):
			continue
		case errors.Is(err, io.EOF):
			return fmt.Errorf("excessive line got %w", io.ErrUnexpectedEOF)
		default:
			return fmt.Errorf("excessive line stranded: %w", err)
		}
	}
}

// IsFromLine returns whether line matches the mbox header pattern.
func IsFromLine(line []byte) bool {
	switch {
	case len(line) < 12,
		line[0] != 'F',
		line[1] != 'r',
		line[2] != 'o',
		line[3] != 'm',
		line[4] != ' ',
		line[len(line)-7] != ' ',
		line[len(line)-2] != '\r',
		line[len(line)-1] != '\n':
		return false
	}

	// year should be ASCII decimal
	m := line[len(line)-6]
	c := line[len(line)-5]
	d := line[len(line)-4]
	y := line[len(line)-3]
	switch {
	case
		m < '0' || m > '9',
		c < '0' || c > '9',
		d < '0' || d > '9',
		y < '0' || y > '9':
		return false
	}

	addrLen := bytes.IndexByte(line[5:], ' ')
	_, err := mail.ParseAddress(string(line[5 : 5+addrLen]))
	return err == nil
}
