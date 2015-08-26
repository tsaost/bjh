/*
Package bjh is a  Go implementation of Bob Jenkins's byte array hash:
http://burtleburtle.net/bob/hash/doobs.html

Copyright 2015 Sheng-Te Tsao

Use of this source code is governed by the same BSD-style
license that is used by the Go standard library

-------------------------------------------------------------------------------
Based on http://burtleburtle.net/bob/c/lookup3.c,
by Bob Jenkins, May 2006, which is public domain.

Most of the comment are taken from lookup3.c, so "I" refers to
bob Jenkins, not the author of this Go implementation

If you want to find a hash of, say, exactly 7 integers, do

  a = i1;  b = i2;  c = i3;
  mix(a,b,c);
  a += i4; b += i5; c += i6;
  mix(a,b,c);
  a += i7;
  final(a,b,c);

then use c as the hash value. */
package bjh

import (
	"hash"
	"unsafe"
)

type digest struct {
	initval0, initval1 uint32 
	a, b, c uint32
}

// New return NewBobJenkinsHash(0, 0)
func New() hash.Hash32 {
	return NewBJH(0, 0)
}

// NewBJH return a hash.Hash32 that can be used to
// compute Bob Jenkin's Hash
func NewBJH(initval0, initval1 uint32) hash.Hash32 {
	d := new(digest)
	d.initval0, d.initval1 = initval0, initval1
	d.Reset()
	return d
}

// Sum appends the current hash to b and returns the resulting slice.
// It does not change the underlying hash state.
func (d *digest) Sum(b []byte) []byte {
	s := d.Sum32()
	b = append(b, byte(s >> 24))
	b = append(b, byte(s >> 16))
	b = append(b, byte(s >> 8))
	b = append(b, byte(s))
    return b
}

// Size is the number of bytes returned by Bob Jenkin's hash
const Size = 4

// BlockSize returns the hash's underlying block size.
// The Write method must be able to accept any amount
// of data, but it may operate more efficiently if all writes
// are a multiple of the block size.
// http://golang.org/pkg/hash/
func (d *digest) BlockSize() int { return Size * 3 }

func (d *digest) Size() int		 { return Size }

func (d *digest) Sum32() uint32	 { return d.c }


func (d *digest) Write(data []byte) (int, error) {
	d.a, d.b, d.c = Update(data, d.a, d.b, d.c)
	return len(data), nil
}

func (d *digest) Reset() {
	d.a = 0xdeadbeef + d.initval0;
	d.b, d.c = d.a, d.a + d.initval1;
}

// CheckSum2 returns two 32-bit hash values instead of just one.
// This is good enough for hash table lookup with 2^^64 buckets,
// or if you want a second hash if you're not happy with the first,
// or if you want a probably-unique 64-bit ID for the key.
// c is better mixed than b, so use c first.  If you want
// a 64-bit value do something like "c + (((uint64_t)b)<<32)".
func CheckSum2(data []byte, initval0, initval1 uint32) (uint32, uint32) {
	// Similar to hashlittle2 in lookup3.c
	// but len(data) is not used in the intial setup so that the value will be
	// the same as using New() and Write() on the same data
	a := 0xdeadbeef + initval0;
	b, c := a, a + initval1;
	_, b, c = Update(data, a, b, c)
	return c, b
}

// CheckSum1 returns one 32-bit number based on Bob Jenkins' Hash
func CheckSum1(data []byte, initval0, initval1 uint32) uint32 {
	a := 0xdeadbeef + initval0;
	b, c := a, a + initval1;
	_, _, c = Update(data, a, b, c)
	return c
}

// CheckSum returns one 32-bit number based on Bob Jenkins' Hash
// with initval0 and initval1 both set to zero
func CheckSum(data []byte) uint32 {
	_, _, c := Update(data, 0xdeadbeef, 0xdeadbeef, 0xdeadbeef)
	return c
}


// Update takes a slice of bytes and update the 3 accumulars a, b, and c
// 
// This is Go implementation that process the data one byte at a time
// so it will work on all machine
/*
func Update(data []byte, a, b, c uint32) (uint32, uint32, uint32) {
	lenData := len(data)
	if lenData == 0 {
		return a, b, c
	}
	remain := lenData
	// 3 uint32 at a time (3 * 4 = 12)
	for i := 0; remain > 12; remain -= 12 {
		val := uint32(data[i]);       i++ // 0-4 bytes to a little endian
		val |= uint32(data[i]) << 8;  i++
		val |= uint32(data[i]) << 16; i++
		val |= uint32(data[i]) << 24; i++
		a += val	
		val  = uint32(data[i]);       i++ // 0-4 bytes to a little endian
		val |= uint32(data[i]) << 8;  i++
		val |= uint32(data[i]) << 16; i++
		val |= uint32(data[i]) << 24; i++
		b += val
		val  = uint32(data[i]);       i++ // 0-4 bytes to a little endian
		val |= uint32(data[i]) << 8;  i++
		val |= uint32(data[i]) << 16; i++
		val |= uint32(data[i]) << 24; i++
		c += val

		// mix -- mix 3 32-bit values reversibly.
		//
        // This is reversible, so any information in (a,b,c) before mix()
        // is still in (a,b,c) after mix().
		//
        // If four pairs of (a,b,c) inputs are run through mix(),
		// or through mix() in reverse, there are at least 32 bits
		// of the output that are sometimes the same for one pair
		// and different for another pair.
		//
        // This was tested for:
		//
        // 	* pairs that differed by one bit, by two bits, in any combination 
		//	  of top bits of (a,b,c), or in any combination of
        // 	  bottom bits of (a,b,c)
		//
        // 	* "differ" is defined as +, -, ^, or ~^.  For + and -, I
        // 	  transformed the output delta to a Gray code (a^(a>>1)) so a
        // 	  string of 1's (as is commonly produced by subtraction) look
        // 	  like a single 1-bit difference.
		//
        // 	* the base values were pseudorandom, all zero but one bit set,
        // 	  or all zero plus a counter that starts at zero.
		//
        // Some k values for my "a-=c; a^=rot(c,k); c+=b;" arrangement
		// that satisfy this are
		//   4  6  8 16 19  4
		//   9 15  3 18 27 15
		//  14  9  3  7 17  3
        // Well, "9 15 3 18 27 15" didn't quite get 32 bits diffing for
        // "differ" defined as + with a one-bit base and a two-bit delta.
        // I used http://burtleburtle.net/bob/hash/avalanche.html to choose
        // the operations, constants, and arrangements of the variables.
		//
        // This does not achieve avalanche.  There are input bits of
        // (a,b,c) that fail to affect some output bits of (a,b,c),
        // especially of a.  The most thoroughly mixed value is c, but it
        // doesn't really even achieve avalanche in c.
		//
		// This allows some parallelism.  Read-after-writes are good at
		// doubling the number of bits affected, so the goal of mixing
		// pulls in the opposite direction as the goal of parallelism.
		// I did what I could.
		//
		// Rotates seem to cost as much as shifts on every machine I could
		// lay my hands on, and rotates are much kinder to the top and
		// bottom bits, so I used rotates.
		// -----------------------------------------------------------------
		// #define rot(x,k) (((x)<<(k)) | ((x)>>(32-(k))))
		// #define mix(a,b,c) { \
		// a -= c;  a ^= rot(c, 4);  c += b; \
		// b -= a;  b ^= rot(a, 6);  a += c; \
		// c -= b;  c ^= rot(b, 8);  b += a; \
		// a -= c;  a ^= rot(c,16);  c += b; \
		// b -= a;  b ^= rot(a,19);  a += c; \
		// c -= b;  c ^= rot(b, 4);  b += a; }
		a -= c; a ^= c<< 4 | c>>(32	- 4);  c += b
		b -= a; b ^= a<< 6 | a>>(32	- 6);  a += c
		c -= b; c ^= b<< 8 | b>>(32	- 8);  b += a
		a -= c; a ^= c<<16 | c>>(32	- 16); c += b
		b -= a; b ^= a<<19 | a>>(32	- 19); a += c
		c -= b; c ^= b<< 4 | b>>(32 - 4);  b += a
	}


    // -------------------------------- last block: affect all 32 bits of (c)
	data = data[lenData - remain:]
    switch(remain) {
    case 12: c += uint32(data[11])<<24; fallthrough
    case 11: c += uint32(data[10])<<16; fallthrough
    case 10: c += uint32(data[9])<<8;	fallthrough
    case 9 : c += uint32(data[8]);		fallthrough
    case 8 : b += uint32(data[7])<<24;	fallthrough
    case 7 : b += uint32(data[6])<<16;	fallthrough
    case 6 : b += uint32(data[5])<<8;	fallthrough
    case 5 : b += uint32(data[4]);		fallthrough
    case 4 : a += uint32(data[3])<<24;	fallthrough
    case 3 : a += uint32(data[2])<<16;	fallthrough
    case 2 : a += uint32(data[1])<<8;	fallthrough
    case 1 : a += uint32(data[0]);
		//  final mixing of 3 32-bit values (a,b,c) into c
		//
		// Pairs of (a,b,c) values differing in only a few bits will usually
		// produce values of c that look totally different.  This was tested for
		//
		// * pairs that differed by one bit, by two bits, in any combination
		//   of top bits of (a,b,c), or in any combination of
		//	 bottom bits of (a,b,c).
		// 
		// * "differ" is defined as +, -, ^, or ~^.  For + and -, I transformed
		//	 the output delta to a Gray code (a^(a>>1)) so a string of 1's
		//	 (as is commonly produced by subtraction) look like a single
		//	 1-bit difference.
		//
		// * the base values were pseudorandom, all zero but one bit set,
		//   or all zero plus a counter that starts at zero.
		//
		// These constants passed:
		//   14 11 25 16 4 14 24
		//	 12 14 25 16 4 14 24
		// and these came close:
		//   4  8 15 26 3 22 24
		// 10  8 15 26 3 22 24
		// 11  8 15 26 3 22 24
		// --------------------------------------------------------------------
		// #define final(a,b,c) { \
		// c ^= b; c -= rot(b,14); \
		// a ^= c; a -= rot(c,11); \
		// b ^= a; b -= rot(a,25); \
		// c ^= b; c -= rot(b,16); \
		// a ^= c; a -= rot(c,4);  \
		// b ^= a; b -= rot(a,14); \
		// c ^= b; c -= rot(b,24); }
		c ^= b; c -= b<<14 | b>>(32-14)
		a ^= c; a -= c<<11 | c>>(32-11)
		b ^= a; b -= a<<25 | a>>(32-25)
		c ^= b; c -= b<<16 | b>>(32-16)
		a ^= c; a -= c<< 4 | c>>(32-4)
		b ^= a; b -= a<<14 | a>>(32-14)
		c ^= b; c -= b<<24 | b>>(32-24)
	}
	return a, b, c
}
*/

// Update takes a slice of bytes and update the 3 accumulars a, b, and c
// 
// Faster version which reads bytes as unsigned 32 bit little endian integers
// (so it may not work on non-x86) using Unsafe.Pointer
func Update(data []byte, a, b, c uint32) (uint32, uint32, uint32) {
	lenData := len(data)
	if lenData == 0 {
		return a, b, c
	}
	remain := lenData

	k := uintptr(unsafe.Pointer(&data[0]))
	for ; remain > 12; remain -= 12 {
		a += *(*uint32)(unsafe.Pointer(k))
		b += *(*uint32)(unsafe.Pointer(k+4))
		c += *(*uint32)(unsafe.Pointer(k+8))
		k += 12
		// mix -- mix 3 32-bit values reversibly.
		//
        // This is reversible, so any information in (a,b,c) before mix()
        // is still in (a,b,c) after mix().
		//
        // If four pairs of (a,b,c) inputs are run through mix(),
		// or through mix() in reverse, there are at least 32 bits
		// of the output that are sometimes the same for one pair
		// and different for another pair.
		//
        // This was tested for:
		//
        // 	* pairs that differed by one bit, by two bits, in any combination 
		//	  of top bits of (a,b,c), or in any combination of
        // 	  bottom bits of (a,b,c)
		//
        // 	* "differ" is defined as +, -, ^, or ~^.  For + and -, I
        // 	  transformed the output delta to a Gray code (a^(a>>1)) so a
        // 	  string of 1's (as is commonly produced by subtraction) look
        // 	  like a single 1-bit difference.
		//
        // 	* the base values were pseudorandom, all zero but one bit set,
        // 	  or all zero plus a counter that starts at zero.
		//
        // Some k values for my "a-=c; a^=rot(c,k); c+=b;" arrangement
		// that satisfy this are
		//   4  6  8 16 19  4
		//   9 15  3 18 27 15
		//  14  9  3  7 17  3
        // Well, "9 15 3 18 27 15" didn't quite get 32 bits diffing for
        // "differ" defined as + with a one-bit base and a two-bit delta.
        // I used http://burtleburtle.net/bob/hash/avalanche.html to choose
        // the operations, constants, and arrangements of the variables.
		//
        // This does not achieve avalanche.  There are input bits of
        // (a,b,c) that fail to affect some output bits of (a,b,c),
        // especially of a.  The most thoroughly mixed value is c, but it
        // doesn't really even achieve avalanche in c.
		//
		// This allows some parallelism.  Read-after-writes are good at
		// doubling the number of bits affected, so the goal of mixing
		// pulls in the opposite direction as the goal of parallelism.
		// I did what I could.
		//
		// Rotates seem to cost as much as shifts on every machine I could
		// lay my hands on, and rotates are much kinder to the top and
		// bottom bits, so I used rotates.
		// -----------------------------------------------------------------
		// #define rot(x,k) (((x)<<(k)) | ((x)>>(32-(k))))
		// #define mix(a,b,c) { \
		// a -= c;  a ^= rot(c, 4);  c += b; \
		// b -= a;  b ^= rot(a, 6);  a += c; \
		// c -= b;  c ^= rot(b, 8);  b += a; \
		// a -= c;  a ^= rot(c,16);  c += b; \
		// b -= a;  b ^= rot(a,19);  a += c; \
		// c -= b;  c ^= rot(b, 4);  b += a; }
		a -= c; a ^= c<< 4 | c>>(32	- 4);  c += b
		b -= a; b ^= a<< 6 | a>>(32	- 6);  a += c
		c -= b; c ^= b<< 8 | b>>(32	- 8);  b += a
		a -= c; a ^= c<<16 | c>>(32	- 16); c += b
		b -= a; b ^= a<<19 | a>>(32	- 19); a += c
		c -= b; c ^= b<< 4 | b>>(32 - 4);  b += a
	}


    // -------------------------------- last block: affect all 32 bits of (c)
	data = data[lenData - remain:]
    switch(remain) {
    case 12: c += uint32(data[11])<<24; fallthrough
    case 11: c += uint32(data[10])<<16; fallthrough
    case 10: c += uint32(data[9])<<8;	fallthrough
    case 9 : c += uint32(data[8]);		fallthrough
    case 8 : b += uint32(data[7])<<24;	fallthrough
    case 7 : b += uint32(data[6])<<16;	fallthrough
    case 6 : b += uint32(data[5])<<8;	fallthrough
    case 5 : b += uint32(data[4]);		fallthrough
    case 4 : a += uint32(data[3])<<24;	fallthrough
    case 3 : a += uint32(data[2])<<16;	fallthrough
    case 2 : a += uint32(data[1])<<8;	fallthrough
    case 1 : a += uint32(data[0]);
		//  final mixing of 3 32-bit values (a,b,c) into c
		//
		// Pairs of (a,b,c) values differing in only a few bits will usually
		// produce values of c that look totally different.  This was tested for
		//
		// * pairs that differed by one bit, by two bits, in any combination
		//   of top bits of (a,b,c), or in any combination of
		//	 bottom bits of (a,b,c).
		// 
		// * "differ" is defined as +, -, ^, or ~^.  For + and -, I transformed
		//	 the output delta to a Gray code (a^(a>>1)) so a string of 1's
		//	 (as is commonly produced by subtraction) look like a single
		//	 1-bit difference.
		//
		// * the base values were pseudorandom, all zero but one bit set,
		//   or all zero plus a counter that starts at zero.
		//
		// These constants passed:
		//   14 11 25 16 4 14 24
		//	 12 14 25 16 4 14 24
		// and these came close:
		//   4  8 15 26 3 22 24
		// 10  8 15 26 3 22 24
		// 11  8 15 26 3 22 24
		// --------------------------------------------------------------------
		// #define final(a,b,c) { \
		// c ^= b; c -= rot(b,14); \
		// a ^= c; a -= rot(c,11); \
		// b ^= a; b -= rot(a,25); \
		// c ^= b; c -= rot(b,16); \
		// a ^= c; a -= rot(c,4);  \
		// b ^= a; b -= rot(a,14); \
		// c ^= b; c -= rot(b,24); }
		c ^= b; c -= b<<14 | b>>(32-14)
		a ^= c; a -= c<<11 | c>>(32-11)
		b ^= a; b -= a<<25 | a>>(32-25)
		c ^= b; c -= b<<16 | b>>(32-16)
		a ^= c; a -= c<< 4 | c>>(32-4)
		b ^= a; b -= a<<14 | a>>(32-14)
		c ^= b; c -= b<<24 | b>>(32-24)
	}
	return a, b, c
}
