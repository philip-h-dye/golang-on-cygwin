// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package block

import (
	"bytes";
	"crypto/aes";
	"fmt";
	"io";
	"testing";
)

// Test vectors from http://www.cs.ucdavis.edu/~rogaway/papers/eax.pdf

type eaxAESTest struct {
	msg	[]byte;
	key	[]byte;
	nonce	[]byte;
	header	[]byte;
	cipher	[]byte;
}

var eaxAESTests = []eaxAESTest{
	eaxAESTest{
		[]byte{},
		[]byte{0x23, 0x39, 0x52, 0xDE, 0xE4, 0xD5, 0xED, 0x5F, 0x9B, 0x9C, 0x6D, 0x6F, 0xF8, 0x0F, 0xF4, 0x78},
		[]byte{0x62, 0xEC, 0x67, 0xF9, 0xC3, 0xA4, 0xA4, 0x07, 0xFC, 0xB2, 0xA8, 0xC4, 0x90, 0x31, 0xA8, 0xB3},
		[]byte{0x6B, 0xFB, 0x91, 0x4F, 0xD0, 0x7E, 0xAE, 0x6B},
		[]byte{0xE0, 0x37, 0x83, 0x0E, 0x83, 0x89, 0xF2, 0x7B, 0x02, 0x5A, 0x2D, 0x65, 0x27, 0xE7, 0x9D, 0x01},
	},
	eaxAESTest{
		[]byte{0xF7, 0xFB},
		[]byte{0x91, 0x94, 0x5D, 0x3F, 0x4D, 0xCB, 0xEE, 0x0B, 0xF4, 0x5E, 0xF5, 0x22, 0x55, 0xF0, 0x95, 0xA4},
		[]byte{0xBE, 0xCA, 0xF0, 0x43, 0xB0, 0xA2, 0x3D, 0x84, 0x31, 0x94, 0xBA, 0x97, 0x2C, 0x66, 0xDE, 0xBD},
		[]byte{0xFA, 0x3B, 0xFD, 0x48, 0x06, 0xEB, 0x53, 0xFA},
		[]byte{0x19, 0xDD, 0x5C, 0x4C, 0x93, 0x31, 0x04, 0x9D, 0x0B, 0xDA, 0xB0, 0x27, 0x74, 0x08, 0xF6, 0x79, 0x67, 0xE5},
	},
	eaxAESTest{
		[]byte{0x1A, 0x47, 0xCB, 0x49, 0x33},
		[]byte{0x01, 0xF7, 0x4A, 0xD6, 0x40, 0x77, 0xF2, 0xE7, 0x04, 0xC0, 0xF6, 0x0A, 0xDA, 0x3D, 0xD5, 0x23},
		[]byte{0x70, 0xC3, 0xDB, 0x4F, 0x0D, 0x26, 0x36, 0x84, 0x00, 0xA1, 0x0E, 0xD0, 0x5D, 0x2B, 0xFF, 0x5E},
		[]byte{0x23, 0x4A, 0x34, 0x63, 0xC1, 0x26, 0x4A, 0xC6},
		[]byte{0xD8, 0x51, 0xD5, 0xBA, 0xE0, 0x3A, 0x59, 0xF2, 0x38, 0xA2, 0x3E, 0x39, 0x19, 0x9D, 0xC9, 0x26, 0x66, 0x26, 0xC4, 0x0F, 0x80},
	},
	eaxAESTest{
		[]byte{0x48, 0x1C, 0x9E, 0x39, 0xB1},
		[]byte{0xD0, 0x7C, 0xF6, 0xCB, 0xB7, 0xF3, 0x13, 0xBD, 0xDE, 0x66, 0xB7, 0x27, 0xAF, 0xD3, 0xC5, 0xE8},
		[]byte{0x84, 0x08, 0xDF, 0xFF, 0x3C, 0x1A, 0x2B, 0x12, 0x92, 0xDC, 0x19, 0x9E, 0x46, 0xB7, 0xD6, 0x17},
		[]byte{0x33, 0xCC, 0xE2, 0xEA, 0xBF, 0xF5, 0xA7, 0x9D},
		[]byte{0x63, 0x2A, 0x9D, 0x13, 0x1A, 0xD4, 0xC1, 0x68, 0xA4, 0x22, 0x5D, 0x8E, 0x1F, 0xF7, 0x55, 0x93, 0x99, 0x74, 0xA7, 0xBE, 0xDE},
	},
	eaxAESTest{
		[]byte{0x40, 0xD0, 0xC0, 0x7D, 0xA5, 0xE4},
		[]byte{0x35, 0xB6, 0xD0, 0x58, 0x00, 0x05, 0xBB, 0xC1, 0x2B, 0x05, 0x87, 0x12, 0x45, 0x57, 0xD2, 0xC2},
		[]byte{0xFD, 0xB6, 0xB0, 0x66, 0x76, 0xEE, 0xDC, 0x5C, 0x61, 0xD7, 0x42, 0x76, 0xE1, 0xF8, 0xE8, 0x16},
		[]byte{0xAE, 0xB9, 0x6E, 0xAE, 0xBE, 0x29, 0x70, 0xE9},
		[]byte{0x07, 0x1D, 0xFE, 0x16, 0xC6, 0x75, 0xCB, 0x06, 0x77, 0xE5, 0x36, 0xF7, 0x3A, 0xFE, 0x6A, 0x14, 0xB7, 0x4E, 0xE4, 0x98, 0x44, 0xDD},
	},
	eaxAESTest{
		[]byte{0x4D, 0xE3, 0xB3, 0x5C, 0x3F, 0xC0, 0x39, 0x24, 0x5B, 0xD1, 0xFB, 0x7D},
		[]byte{0xBD, 0x8E, 0x6E, 0x11, 0x47, 0x5E, 0x60, 0xB2, 0x68, 0x78, 0x4C, 0x38, 0xC6, 0x2F, 0xEB, 0x22},
		[]byte{0x6E, 0xAC, 0x5C, 0x93, 0x07, 0x2D, 0x8E, 0x85, 0x13, 0xF7, 0x50, 0x93, 0x5E, 0x46, 0xDA, 0x1B},
		[]byte{0xD4, 0x48, 0x2D, 0x1C, 0xA7, 0x8D, 0xCE, 0x0F},
		[]byte{0x83, 0x5B, 0xB4, 0xF1, 0x5D, 0x74, 0x3E, 0x35, 0x0E, 0x72, 0x84, 0x14, 0xAB, 0xB8, 0x64, 0x4F, 0xD6, 0xCC, 0xB8, 0x69, 0x47, 0xC5, 0xE1, 0x05, 0x90, 0x21, 0x0A, 0x4F},
	},
	eaxAESTest{
		[]byte{0x8B, 0x0A, 0x79, 0x30, 0x6C, 0x9C, 0xE7, 0xED, 0x99, 0xDA, 0xE4, 0xF8, 0x7F, 0x8D, 0xD6, 0x16, 0x36},
		[]byte{0x7C, 0x77, 0xD6, 0xE8, 0x13, 0xBE, 0xD5, 0xAC, 0x98, 0xBA, 0xA4, 0x17, 0x47, 0x7A, 0x2E, 0x7D},
		[]byte{0x1A, 0x8C, 0x98, 0xDC, 0xD7, 0x3D, 0x38, 0x39, 0x3B, 0x2B, 0xF1, 0x56, 0x9D, 0xEE, 0xFC, 0x19},
		[]byte{0x65, 0xD2, 0x01, 0x79, 0x90, 0xD6, 0x25, 0x28},
		[]byte{0x02, 0x08, 0x3E, 0x39, 0x79, 0xDA, 0x01, 0x48, 0x12, 0xF5, 0x9F, 0x11, 0xD5, 0x26, 0x30, 0xDA, 0x30, 0x13, 0x73, 0x27, 0xD1, 0x06, 0x49, 0xB0, 0xAA, 0x6E, 0x1C, 0x18, 0x1D, 0xB6, 0x17, 0xD7, 0xF2},
	},
	eaxAESTest{
		[]byte{0x1B, 0xDA, 0x12, 0x2B, 0xCE, 0x8A, 0x8D, 0xBA, 0xF1, 0x87, 0x7D, 0x96, 0x2B, 0x85, 0x92, 0xDD, 0x2D, 0x56},
		[]byte{0x5F, 0xFF, 0x20, 0xCA, 0xFA, 0xB1, 0x19, 0xCA, 0x2F, 0xC7, 0x35, 0x49, 0xE2, 0x0F, 0x5B, 0x0D},
		[]byte{0xDD, 0xE5, 0x9B, 0x97, 0xD7, 0x22, 0x15, 0x6D, 0x4D, 0x9A, 0xFF, 0x2B, 0xC7, 0x55, 0x98, 0x26},
		[]byte{0x54, 0xB9, 0xF0, 0x4E, 0x6A, 0x09, 0x18, 0x9A},
		[]byte{0x2E, 0xC4, 0x7B, 0x2C, 0x49, 0x54, 0xA4, 0x89, 0xAF, 0xC7, 0xBA, 0x48, 0x97, 0xED, 0xCD, 0xAE, 0x8C, 0xC3, 0x3B, 0x60, 0x45, 0x05, 0x99, 0xBD, 0x02, 0xC9, 0x63, 0x82, 0x90, 0x2A, 0xEF, 0x7F, 0x83, 0x2A},
	},
	eaxAESTest{
		[]byte{0x6C, 0xF3, 0x67, 0x20, 0x87, 0x2B, 0x85, 0x13, 0xF6, 0xEA, 0xB1, 0xA8, 0xA4, 0x44, 0x38, 0xD5, 0xEF, 0x11},
		[]byte{0xA4, 0xA4, 0x78, 0x2B, 0xCF, 0xFD, 0x3E, 0xC5, 0xE7, 0xEF, 0x6D, 0x8C, 0x34, 0xA5, 0x61, 0x23},
		[]byte{0xB7, 0x81, 0xFC, 0xF2, 0xF7, 0x5F, 0xA5, 0xA8, 0xDE, 0x97, 0xA9, 0xCA, 0x48, 0xE5, 0x22, 0xEC},
		[]byte{0x89, 0x9A, 0x17, 0x58, 0x97, 0x56, 0x1D, 0x7E},
		[]byte{0x0D, 0xE1, 0x8F, 0xD0, 0xFD, 0xD9, 0x1E, 0x7A, 0xF1, 0x9F, 0x1D, 0x8E, 0xE8, 0x73, 0x39, 0x38, 0xB1, 0xE8, 0xE7, 0xF6, 0xD2, 0x23, 0x16, 0x18, 0x10, 0x2F, 0xDB, 0x7F, 0xE5, 0x5F, 0xF1, 0x99, 0x17, 0x00},
	},
	eaxAESTest{
		[]byte{0xCA, 0x40, 0xD7, 0x44, 0x6E, 0x54, 0x5F, 0xFA, 0xED, 0x3B, 0xD1, 0x2A, 0x74, 0x0A, 0x65, 0x9F, 0xFB, 0xBB, 0x3C, 0xEA, 0xB7},
		[]byte{0x83, 0x95, 0xFC, 0xF1, 0xE9, 0x5B, 0xEB, 0xD6, 0x97, 0xBD, 0x01, 0x0B, 0xC7, 0x66, 0xAA, 0xC3},
		[]byte{0x22, 0xE7, 0xAD, 0xD9, 0x3C, 0xFC, 0x63, 0x93, 0xC5, 0x7E, 0xC0, 0xB3, 0xC1, 0x7D, 0x6B, 0x44},
		[]byte{0x12, 0x67, 0x35, 0xFC, 0xC3, 0x20, 0xD2, 0x5A},
		[]byte{0xCB, 0x89, 0x20, 0xF8, 0x7A, 0x6C, 0x75, 0xCF, 0xF3, 0x96, 0x27, 0xB5, 0x6E, 0x3E, 0xD1, 0x97, 0xC5, 0x52, 0xD2, 0x95, 0xA7, 0xCF, 0xC4, 0x6A, 0xFC, 0x25, 0x3B, 0x46, 0x52, 0xB1, 0xAF, 0x37, 0x95, 0xB1, 0x24, 0xAB, 0x6E},
	},
}

func TestEAXEncrypt_AES(t *testing.T) {
	b := new(bytes.Buffer);
	for i, tt := range eaxAESTests {
		test := fmt.Sprintf("test %d", i);
		c, err := aes.NewCipher(tt.key);
		if err != nil {
			t.Fatalf("%s: NewCipher(%d bytes) = %s", test, len(tt.key), err)
		}
		b.Reset();
		enc := NewEAXEncrypter(c, tt.nonce, tt.header, 16, b);
		n, err := io.Copy(enc, bytes.NewBuffer(tt.msg));
		if n != int64(len(tt.msg)) || err != nil {
			t.Fatalf("%s: io.Copy into encrypter: %d, %s", test, n, err)
		}
		err = enc.Close();
		if err != nil {
			t.Fatalf("%s: enc.Close: %s", test, err)
		}
		if d := b.Bytes(); !same(d, tt.cipher) {
			t.Fatalf("%s: got %x want %x", test, d, tt.cipher)
		}
	}
}

func TestEAXDecrypt_AES(t *testing.T) {
	b := new(bytes.Buffer);
	for i, tt := range eaxAESTests {
		test := fmt.Sprintf("test %d", i);
		c, err := aes.NewCipher(tt.key);
		if err != nil {
			t.Fatalf("%s: NewCipher(%d bytes) = %s", test, len(tt.key), err)
		}
		b.Reset();
		dec := NewEAXDecrypter(c, tt.nonce, tt.header, 16, bytes.NewBuffer(tt.cipher));
		n, err := io.Copy(b, dec);
		if n != int64(len(tt.msg)) || err != nil {
			t.Fatalf("%s: io.Copy into decrypter: %d, %s", test, n, err)
		}
		if d := b.Bytes(); !same(d, tt.msg) {
			t.Fatalf("%s: got %x want %x", test, d, tt.msg)
		}
	}
}
