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
package cyclic_test

import (
	"crypto/rand"
	"fmt"
	"github.com/karalabe/iris/crypto/cyclic"
)

func Example_usage() {
	// Generate the cyclic group
	group, err := cyclic.New(rand.Reader, 120)
	if err != nil {
		fmt.Println("Failed to generate cyclic group:", err)
	}
	// Output in a nice, source friendly byte format
	fmt.Println("Cyclic group base:")
	bytes := group.Base.Bytes()
	for byte := 0; byte < len(bytes); byte++ {
		fmt.Printf("0x%02x, ", bytes[byte])
		if byte%8 == 7 {
			fmt.Println()
		}
	}
	fmt.Println()
	// Output in a nice, source friendly byte format
	fmt.Println("Cyclic group generator:")
	bytes = group.Generator.Bytes()
	for byte := 0; byte < len(bytes); byte++ {
		fmt.Printf("0x%02x, ", bytes[byte])
		if byte%8 == 7 {
			fmt.Println()
		}
	}
	fmt.Println()
}
