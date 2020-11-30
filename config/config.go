/*
 * ECAL
 *
 * Copyright 2020 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the MIT
 * License, If a copy of the MIT License was not distributed with this
 * file, You can obtain one at https://opensource.org/licenses/MIT.
 */

package config

import (
	"fmt"
	"strconv"

	"devt.de/krotik/common/errorutil"
)

// Global variables
// ================

/*
ProductVersion is the current version of ECAL
*/
const ProductVersion = "1.0.0"

/*
Known configuration options for ECAL
*/
const (
	WorkerCount = "WorkerCount"
)

/*
DefaultConfig is the defaut configuration
*/
var DefaultConfig = map[string]interface{}{
	WorkerCount: 1,
}

/*
Config is the actual config which is used
*/
var Config map[string]interface{}

/*
Initialise the config
*/
func init() {
	data := make(map[string]interface{})
	for k, v := range DefaultConfig {
		data[k] = v
	}

	Config = data
}

// Helper functions
// ================

/*
Str reads a config value as a string value.
*/
func Str(key string) string {
	return fmt.Sprint(Config[key])
}

/*
Int reads a config value as an int value.
*/
func Int(key string) int {
	ret, err := strconv.ParseInt(fmt.Sprint(Config[key]), 10, 64)

	errorutil.AssertTrue(err == nil,
		fmt.Sprintf("Could not parse config key %v: %v", key, err))

	return int(ret)
}

/*
Bool reads a config value as a boolean value.
*/
func Bool(key string) bool {
	ret, err := strconv.ParseBool(fmt.Sprint(Config[key]))

	errorutil.AssertTrue(err == nil,
		fmt.Sprintf("Could not parse config key %v: %v", key, err))

	return ret
}
