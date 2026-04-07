// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

// IDemoHelloService Define Demo operator interface
type IDemoHelloService interface {
	Hello() (result string)
}
