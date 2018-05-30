package ecoji

import (
	"io"
)

type RuneWriter interface {
	WriteByte(byte) error
	WriteRune(rune) (int, error)
}

func encode(s []byte, w RuneWriter) (err error) {

	if len(s) == 0 {
		panic("expected data")
	}

	var b0, b1, b2, b3, b4 int = int(s[0]), 0, 0, 0, 0

	if len(s) > 1 {
		b1 = int(s[1])
	}

	if len(s) > 2 {
		b2 = int(s[2])
	}

	if len(s) > 3 {
		b3 = int(s[3])
	}

	if len(s) > 4 {
		b4 = int(s[4])
	}

	runes := []rune{emojis[b0<<2|b1>>6], padding, padding, padding}

	switch len(s) {
	case 1:
	//nothing to do, all padding
	case 2:
		runes[1] = emojis[(b1&0x3f)<<4|b2>>4]
	case 3:
		runes[1] = emojis[(b1&0x3f)<<4|b2>>4]
		runes[2] = emojis[(b2&0x0f)<<6|b3>>2]
	case 4:
		runes[1] = emojis[(b1&0x3f)<<4|b2>>4]
		runes[2] = emojis[(b2&0x0f)<<6|b3>>2]

		switch b3 & 0x03 {
		case 0:
			runes[3] = padding40
		case 1:
			runes[3] = padding41
		case 2:
			runes[3] = padding42
		case 3:
			runes[3] = padding43
		}

	case 5:
		runes[1] = emojis[(b1&0x3f)<<4|b2>>4]
		runes[2] = emojis[(b2&0x0f)<<6|b3>>2]
		runes[3] = emojis[(b3&0x03)<<8|b4]
	default:
		panic("unexpected length " + string(len(s)))

	}

	for _, r := range runes {
		if _, err := w.WriteRune(r); err != nil {
			return err
		}
	}

	return nil
}

func readFully(r io.Reader, buffer []byte) (n int, e error) {
	num, err := r.Read(buffer)

	for num < len(buffer) && err != io.EOF && err == nil {
		more, err2 := r.Read(buffer[num:])
		num += more
		err = err2
	}

	return num, err
}

//Maps every 10 bits from the reader to one of 1024 Unicode emojis, writing the emojis.
func Encode(r io.Reader, w RuneWriter, wrap uint) (err error) {

	buffer := make([]byte, 5)
	printed := uint(0)

	for {

		num, err := readFully(r, buffer)

		if num == 0 && err == io.EOF {
			if printed > 0 {
				if err := w.WriteByte('\n'); err != nil {
					return err
				}
			}
			break
		}

		if err != nil && err != io.EOF {
			return err
		}

		if err := encode(buffer[0:num], w); err != nil {
			return err
		}

		if wrap > 0 {
			printed += 4
			if printed >= wrap {
				if err := w.WriteByte('\n'); err != nil {
					return err
				}
				printed = 0
			}
		}

	}

	return nil
}
