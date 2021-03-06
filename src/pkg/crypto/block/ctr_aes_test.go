// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// CTR AES test vectors.

// See U.S. National Institute of Standards and Technology (NIST)
// Special Publication 800-38A, ``Recommendation for Block Cipher
// Modes of Operation,'' 2001 Edition, pp. 55-58.

package block

import (
	"bytes";
	"crypto/aes";
	"io";
	"testing";
)

type ctrTest struct {
	name	string;
	key	[]byte;
	iv	[]byte;
	in	[]byte;
	out	[]byte;
}

var commonCounter = []byte{0xf0, 0xf1, 0xf2, 0xf3, 0xf4, 0xf5, 0xf6, 0xf7, 0xf8, 0xf9, 0xfa, 0xfb, 0xfc, 0xfd, 0xfe, 0xff}

var ctrAESTests = []ctrTest{
	// NIST SP 800-38A pp 55-58
	ctrTest{
		"CTR-AES128",
		commonKey128,
		commonCounter,
		commonInput,
		[]byte{
			0x87, 0x4d, 0x61, 0x91, 0xb6, 0x20, 0xe3, 0x26, 0x1b, 0xef, 0x68, 0x64, 0x99, 0x0d, 0xb6, 0xce,
			0x98, 0x06, 0xf6, 0x6b, 0x79, 0x70, 0xfd, 0xff, 0x86, 0x17, 0x18, 0x7b, 0xb9, 0xff, 0xfd, 0xff,
			0x5a, 0xe4, 0xdf, 0x3e, 0xdb, 0xd5, 0xd3, 0x5e, 0x5b, 0x4f, 0x09, 0x02, 0x0d, 0xb0, 0x3e, 0xab,
			0x1e, 0x03, 0x1d, 0xda, 0x2f, 0xbe, 0x03, 0xd1, 0x79, 0x21, 0x70, 0xa0, 0xf3, 0x00, 0x9c, 0xee,
		},
	},
	ctrTest{
		"CTR-AES192",
		commonKey192,
		commonCounter,
		commonInput,
		[]byte{
			0x1a, 0xbc, 0x93, 0x24, 0x17, 0x52, 0x1c, 0xa2, 0x4f, 0x2b, 0x04, 0x59, 0xfe, 0x7e, 0x6e, 0x0b,
			0x09, 0x03, 0x39, 0xec, 0x0a, 0xa6, 0xfa, 0xef, 0xd5, 0xcc, 0xc2, 0xc6, 0xf4, 0xce, 0x8e, 0x94,
			0x1e, 0x36, 0xb2, 0x6b, 0xd1, 0xeb, 0xc6, 0x70, 0xd1, 0xbd, 0x1d, 0x66, 0x56, 0x20, 0xab, 0xf7,
			0x4f, 0x78, 0xa7, 0xf6, 0xd2, 0x98, 0x09, 0x58, 0x5a, 0x97, 0xda, 0xec, 0x58, 0xc6, 0xb0, 0x50,
		},
	},
	ctrTest{
		"CTR-AES256",
		commonKey256,
		commonCounter,
		commonInput,
		[]byte{
			0x60, 0x1e, 0xc3, 0x13, 0x77, 0x57, 0x89, 0xa5, 0xb7, 0xa7, 0xf5, 0x04, 0xbb, 0xf3, 0xd2, 0x28,
			0xf4, 0x43, 0xe3, 0xca, 0x4d, 0x62, 0xb5, 0x9a, 0xca, 0x84, 0xe9, 0x90, 0xca, 0xca, 0xf5, 0xc5,
			0x2b, 0x09, 0x30, 0xda, 0xa2, 0x3d, 0xe9, 0x4c, 0xe8, 0x70, 0x17, 0xba, 0x2d, 0x84, 0x98, 0x8d,
			0xdf, 0xc9, 0xc5, 0x8d, 0xb6, 0x7a, 0xad, 0xa6, 0x13, 0xc2, 0xdd, 0x08, 0x45, 0x79, 0x41, 0xa6,
		},
	},
}

func TestCTR_AES(t *testing.T) {
	for _, tt := range ctrAESTests {
		test := tt.name;

		c, err := aes.NewCipher(tt.key);
		if err != nil {
			t.Errorf("%s: NewCipher(%d bytes) = %s", test, len(tt.key), err);
			continue;
		}

		for j := 0; j <= 5; j += 5 {
			var crypt bytes.Buffer;
			in := tt.in[0 : len(tt.in)-j];
			w := NewCTRWriter(c, tt.iv, &crypt);
			var r io.Reader = bytes.NewBuffer(in);
			n, err := io.Copy(w, r);
			if n != int64(len(in)) || err != nil {
				t.Errorf("%s/%d: CTRWriter io.Copy = %d, %v want %d, nil", test, len(in), n, err, len(in))
			} else if d, out := crypt.Bytes(), tt.out[0:len(in)]; !same(out, d) {
				t.Errorf("%s/%d: CTRWriter\ninpt %x\nhave %x\nwant %x", test, len(in), in, d, out)
			}
		}

		for j := 0; j <= 7; j += 7 {
			var plain bytes.Buffer;
			out := tt.out[0 : len(tt.out)-j];
			r := NewCTRReader(c, tt.iv, bytes.NewBuffer(out));
			w := &plain;
			n, err := io.Copy(w, r);
			if n != int64(len(out)) || err != nil {
				t.Errorf("%s/%d: CTRReader io.Copy = %d, %v want %d, nil", test, len(out), n, err, len(out))
			} else if d, in := plain.Bytes(), tt.in[0:len(out)]; !same(in, d) {
				t.Errorf("%s/%d: CTRReader\nhave %x\nwant %x", test, len(out), d, in)
			}
		}

		if t.Failed() {
			break
		}
	}
}
