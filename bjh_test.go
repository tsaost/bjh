package bjh

import (
	"testing"
)

type bjhTest struct {
	expected uint32
	data string
}

func Test1(t *testing.T) {
	for _, x := range []bjhTest{
		bjhTest{0xcafebaba, ""},
		bjhTest{0x704dedac, "a"},
		bjhTest{0x48c80e69, "ab"},
		bjhTest{0xd4383038, "abc"},
		bjhTest{0xab6eb87c, "abcd"},
		bjhTest{0x623ed751, "abcde"},
		bjhTest{0x714cb538, "abcdef"},
		bjhTest{0xb2f6ab88, "abcdefg"},
		bjhTest{0x12a39337, "abcdefgh"},
		bjhTest{0x7e6e2145, "abcdefghi"},
		bjhTest{0xc640b76a, "abcdefghij"},
		bjhTest{0x13e5957a, "0123456789a"},
		bjhTest{0x53238445, "0123456789ab"},
		bjhTest{0x0dc6fdc5, "0123456789abc"},
		bjhTest{0xb6b3320c,
			"Premature optimization is the root of all evil - Donald Knuth"},
	} {
		// io.WriteString(c, x.input)
		_, _, c := Update([]byte(x.data), 0xCafeBaba, 0xCafeBaba, 0xCafeBaba)
		if c != x.expected {
			t.Fatalf("bob(%s) = 0x%x want 0x%x", x.data, c, x.expected)
		}
	}
}

type bjhTest2 struct {
	c, b uint32
	data string
	init0, init1 uint32
}

func TestChecksum2(t *testing.T) {
	// These test are from driver5() in lookup3.c
	for _, x := range []bjhTest2{
		bjhTest2{0xdeadbeef, 0xdeadbeef, "", 0, 0},
		bjhTest2{0xbd5b7dde, 0xdeadbeef, "", 0, 0xdeadbeef},
		bjhTest2{0x9c093ccd, 0xbd5b7dde, "", 0xdeadbeef, 0xdeadbeef},
		bjhTest2{0x17770551, 0xce7226e6, "Four score and seven years ago",30,0},
		bjhTest2{0xe3607cae, 0xbd371de4, "Four score and seven years ago",30,1},
		bjhTest2{0xcd628161, 0x6cbea4b3, "Four score and seven years ago",31,0},
	} {
		c, b := CheckSum2([]byte(x.data), x.init0, x.init1)
		if c != x.c {
			t.Fatalf("bob(%s) = c(0x%x) want 0x%x", x.data, c, x.c)
		}
		if b != x.b {
			t.Fatalf("bob(%s) = b(0x%x) want 0x%x", x.data, b, x.b)
		}
	}
}

type bjhTest3 struct {
	c uint32
	data string
	init0, init1 uint32
}

func TestNewBJH(t *testing.T) {
	// These test are from driver5() in lookup3.c
	for _, x := range []bjhTest3{
		bjhTest3{0xdeadbeef, "", 0, 0},
		bjhTest3{0xbd5b7dde, "", 0, 0xdeadbeef},
		bjhTest3{0x9c093ccd, "", 0xdeadbeef, 0xdeadbeef},
		bjhTest3{0x17770551, "Four score and seven years ago",30,0},
		bjhTest3{0xe3607cae, "Four score and seven years ago",30,1},
		bjhTest3{0xcd628161, "Four score and seven years ago",31,0},
	} {
		d := NewBJH(x.init0, x.init1)
		d.Write([]byte(x.data))
		c := d.Sum32()
		if c != x.c {
			t.Fatalf("bob(%s) = c(0x%x) want 0x%x", x.data, c, x.c)
		}

		// test New(), which is equivalent to NewBJH(0, 0)
		d = New()
		d.Write([]byte(""))
		c = d.Sum32()
		if c != 0xdeadbeef {
			t.Fatalf("bob(%s) = c(0x%x) want 0x%x", x.data, c, x.c)
		}

	}
}
