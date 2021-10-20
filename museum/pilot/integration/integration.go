// Copyright 2021 PGHQ. All Rights Reserved.
//
// Licensed under the GNU General Public License, Version 3 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package integration provides resources for doing integration testing.
package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	// NoopRun is a test function for mocking testing.M
	NoopRun = func() int { return 0 }
)

// ExpectExit is a test function for asserting exit codes when exit is called
func ExpectExit(t *testing.T, expect int) func(code int){
	return func(code int) {
		assert.Equal(t, expect, code)
	}
}