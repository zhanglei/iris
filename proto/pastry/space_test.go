// Iris - Decentralized Messaging Framework
// Copyright 2013 Peter Szilagyi. All rights reserved.
//
// Iris is dual licensed: you can redistribute it and/or modify it under the
// terms of the GNU General Public License as published by the Free Software
// Foundation, either version 3 of the License, or (at your option) any later
// version.
//
// The framework is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
// more details.
//
// Alternatively, the Iris framework may be used in accordance with the terms
// and conditions contained in a signed written agreement between you and the
// author(s).
//
// Author: peterke@gmail.com (Peter Szilagyi)

package pastry

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"hash"
	"math/big"
	"testing"

	"github.com/karalabe/iris/config"
)

type spaceTest struct {
	idA    *big.Int
	idB    *big.Int
	delta  *big.Int
	dist   *big.Int
	prefix int
	digit  int
}

var one = big.NewInt(1)

// The tests assume the default 4 bit digits!
var spaceTests = []spaceTest{
	// Simple startup cases
	{big.NewInt(0), big.NewInt(15), big.NewInt(15), big.NewInt(15), config.PastrySpace/config.PastryBase - 1, 15},
	{big.NewInt(15), big.NewInt(0), big.NewInt(-15), big.NewInt(15), config.PastrySpace/config.PastryBase - 1, 0},
	{big.NewInt(0), big.NewInt(127), big.NewInt(127), big.NewInt(127), config.PastrySpace/config.PastryBase - 2, 7},
	{big.NewInt(127), big.NewInt(0), big.NewInt(-127), big.NewInt(127), config.PastrySpace/config.PastryBase - 2, 0},
	{big.NewInt(128), big.NewInt(256), big.NewInt(128), big.NewInt(128), config.PastrySpace/config.PastryBase - 3, 1},
	{big.NewInt(256), big.NewInt(128), big.NewInt(-128), big.NewInt(128), config.PastrySpace/config.PastryBase - 3, 0},

	// Boring cases
	{big.NewInt(65536), big.NewInt(262144), big.NewInt(196608), big.NewInt(196608), config.PastrySpace/config.PastryBase - 5, 4},
	{big.NewInt(262144), big.NewInt(65536), big.NewInt(-196608), big.NewInt(196608), config.PastrySpace/config.PastryBase - 5, 1},

	// Circular wrapping
	{new(big.Int).Sub(modulo, one), big.NewInt(0), big.NewInt(1), big.NewInt(1), 0, 0},
	{big.NewInt(0), new(big.Int).Sub(modulo, one), big.NewInt(-1), big.NewInt(1), 0, 15},
	{new(big.Int).Sub(modulo, one), big.NewInt(1), big.NewInt(2), big.NewInt(2), 0, 0},
	{big.NewInt(1), new(big.Int).Sub(modulo, one), big.NewInt(-2), big.NewInt(2), 0, 15},

	// Half splits
	{big.NewInt(0), posmid, posmid, posmid, 0, 8},
	{posmid, big.NewInt(0), negmid, posmid, 0, 0},
	{big.NewInt(0), new(big.Int).Sub(posmid, one), new(big.Int).Sub(posmid, one), new(big.Int).Sub(posmid, one), 0, 7},
	{new(big.Int).Sub(posmid, one), big.NewInt(0), new(big.Int).Add(negmid, one), new(big.Int).Sub(posmid, one), 0, 0},
	{big.NewInt(0), new(big.Int).Add(posmid, one), new(big.Int).Add(negmid, one), new(big.Int).Sub(posmid, one), 0, 8},
	{new(big.Int).Add(posmid, one), big.NewInt(0), new(big.Int).Sub(posmid, one), new(big.Int).Sub(posmid, one), 0, 0},
}

func TestSpace(t *testing.T) {
	for i, tt := range spaceTests {
		if d := delta(tt.idA, tt.idB); tt.delta.Cmp(d) != 0 {
			t.Errorf("test %d: delta mismatch: have %v, want %v.", i, d, tt.delta)
		}
		if d := Distance(tt.idA, tt.idB); tt.dist.Cmp(d) != 0 {
			t.Errorf("test %d: dist mismatch: have %v, want %v.", i, d, tt.dist)
		}
		if p, d := prefix(tt.idA, tt.idB); tt.prefix != p || tt.digit != d {
			t.Errorf("test %d: prefix/digit mismatch: have %v/%v, want %v/%v.", i, p, d, tt.prefix, tt.digit)
		}
	}
}

type resolveTest struct {
	hasher func() hash.Hash
	bitlen int
	text   string
	id     []byte
}

var resolveTests = []resolveTest{
	// Inter-byte boundaries
	{md5.New, 8, "", []byte{0xd4}},
	{md5.New, 16, "", []byte{0xd4, 0x1d}},
	{md5.New, 24, "", []byte{0xd4, 0x1d, 0x8c}},
	{md5.New, 8, "string", []byte{0xb4}},
	{md5.New, 16, "string", []byte{0xb4, 0x5c}},
	{md5.New, 24, "string", []byte{0xb4, 0x5c, 0xff}},

	// Intra-byte boundaries
	{md5.New, 1, "", []byte{0x00}},
	{md5.New, 2, "", []byte{0x00}},
	{md5.New, 3, "", []byte{0x04}},
	{md5.New, 4, "", []byte{0x04}},
	{md5.New, 5, "", []byte{0x14}},
	{md5.New, 6, "", []byte{0x14}},
	{md5.New, 7, "", []byte{0x54}},

	// Other hashes
	{sha1.New, 32, "", []byte{0xda, 0x39, 0xa3, 0xee}},
	{sha256.New, 32, "", []byte{0xe3, 0xb0, 0xc4, 0x42}},
}

func TestResolve(t *testing.T) {
	// Save the previous config values
	s, h := config.PastrySpace, config.PastryResolver
	defer func() { config.PastrySpace, config.PastryResolver = s, h }()

	// Run the tests
	for i, tt := range resolveTests {
		config.PastrySpace = tt.bitlen
		config.PastryResolver = tt.hasher
		if id := Resolve(tt.text); id.Cmp(new(big.Int).SetBytes(tt.id)) != 0 {
			t.Errorf("test %d: resolution mismatch: have %v, want %v.", i, id.Bytes(), tt.id)
		}
	}
}
