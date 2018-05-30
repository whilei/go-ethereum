package ecoji

import (
	"bytes"
	"strings"
	"testing"
)

func check(t *testing.T, expected []rune, input []byte) {
	reader := bytes.NewBuffer(input)
	buffer1 := bytes.NewBuffer(nil)

	Encode(reader, buffer1, 0)
	actual, _ := buffer1.ReadString('\n')

	if cmp := strings.Compare(actual, string(expected)); cmp != 0 {
		t.Errorf("'%s' != '%s' %d", string(expected), actual, cmp)
	}
}

func TestOneByteEncode(t *testing.T) {
	check(t, []rune{emojis[int('k')<<2], padding, padding, padding}, []byte{'k'})
}

func TestTwoByteEncode(t *testing.T) {
	check(t, []rune{emojis[0], emojis[16], padding, padding}, []byte{0x00, 0x01})
}

func TestThreeByteEncode(t *testing.T) {
	check(t, []rune{emojis[0], emojis[16], emojis[128], padding}, []byte{0x00, 0x01, 0x02})
}

func TestFourByteEncode(t *testing.T) {
	check(t, []rune{emojis[0], emojis[16], emojis[128], padding40}, []byte{0x00, 0x01, 0x02, 0x00})
	check(t, []rune{emojis[0], emojis[16], emojis[128], padding41}, []byte{0x00, 0x01, 0x02, 0x01})
	check(t, []rune{emojis[0], emojis[16], emojis[128], padding42}, []byte{0x00, 0x01, 0x02, 0x02})
	check(t, []rune{emojis[0], emojis[16], emojis[128], padding43}, []byte{0x00, 0x01, 0x02, 0x03})
}

func TestFiveByteEncode(t *testing.T) {
	check(t, []rune{emojis[687], emojis[222], emojis[960], emojis[291]}, []byte{0xab, 0xcd, 0xef, 0x01, 0x23})
}

func TestGarbage(t *testing.T) {
	reader := strings.NewReader("not emojis")
	buffer1 := bytes.NewBuffer(nil)

	err := Decode(reader, buffer1)
	if err == nil {
		t.Error("Expected error")
	}

	if !strings.Contains(err.Error(), "Invalid rune") {
		t.Errorf("Unexpected error message: %s", err.Error())
	}
}
