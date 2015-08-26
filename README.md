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

then use c as the hash value.
