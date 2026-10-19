package arbo

import (
	"encoding/hex"
	"math/big"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestHashSha256(t *testing.T) {
	// Sha256 hash
	hashFunc := &HashSha256{}
	b := []byte("test")
	h, err := hashFunc.Hash(b)
	if err != nil {
		t.Fatal(err)
	}
	c := qt.New(t)
	c.Assert(hex.EncodeToString(h),
		qt.Equals,
		"9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08")
}

func TestHashPoseidon(t *testing.T) {
	// Poseidon hash
	hashFunc := &HashPoseidon{}
	bLen := hashFunc.Len()
	h, err := hashFunc.Hash(
		BigIntToBytes(bLen, big.NewInt(1)),
		BigIntToBytes(bLen, big.NewInt(2)))
	if err != nil {
		t.Fatal(err)
	}
	hBI := BytesToBigInt(h)
	// value checked with circomlib
	c := qt.New(t)
	c.Assert(hBI.String(),
		qt.Equals,
		"7853200120776062878684798364095072458815029376092732009249414926327459813530")
}

func TestHashBlake2b(t *testing.T) {
	// Blake2b hash
	hashFunc := &HashBlake2b{}
	b := []byte("test")
	h, err := hashFunc.Hash(b)
	if err != nil {
		t.Fatal(err)
	}
	c := qt.New(t)
	c.Assert(hex.EncodeToString(h),
		qt.Equals,
		"928b20366943e2afd11ebc0eae2e53a93bf177a4fcf35bcc64d503704e65e202")
}

func TestHashBlake3(t *testing.T) {
	var hashFunc HashFunction = HashFunctionBlake3
	// The first 32 bytes of the official BLAKE3 test vectors:
	// https://github.com/BLAKE3-team/BLAKE3/blob/master/test_vectors/test_vectors.json
	const emptyHash = "af1349b9f5f9a1a6a0404dea36dcc9499bcb25c9adc112b7cc9a93cae41f3262"
	const threeBytesHash = "e1be4d7a8ab5560aa4199eea339849ba8e293d55ca0a81006726d184519e647f"
	tests := []struct {
		name  string
		input [][]byte
		want  string
	}{
		{name: "no inputs", want: emptyHash},
		{name: "empty inputs", input: [][]byte{nil, {}}, want: emptyHash},
		{name: "single input", input: [][]byte{{0, 1, 2}}, want: threeBytesHash},
		{name: "multiple inputs", input: [][]byte{{0}, nil, {1, 2}, {}}, want: threeBytesHash},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			h, err := hashFunc.Hash(test.input...)
			c.Assert(err, qt.IsNil)
			c.Assert(hex.EncodeToString(h), qt.Equals, test.want)
			c.Assert(len(h), qt.Equals, hashFunc.Len())
		})
	}
}
