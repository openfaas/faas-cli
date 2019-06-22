// Copyright (c) OpenFaaS Author(s) 2019. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package flags

import "time"

// TimestampFlag implements the Value interface to accept and validate a
// RFC3339 timestamp string as a flag
type TimestampFlag string

// Type implements pflag.Value
func (t *TimestampFlag) Type() string {
	return "timestamp"
}

// String implements Stringer
func (t *TimestampFlag) String() string {
	if t == nil {
		return ""
	}
	return string(*t)
}

// Set implements pflag.Value
func (t *TimestampFlag) Set(value string) error {
	_, err := time.Parse(time.RFC3339, value)
	if err == nil {
		*t = TimestampFlag(value)
	}
	return err
}

// AsTime returns the underlying time instance
func (t TimestampFlag) AsTime() time.Time {
	v, _ := time.Parse(time.RFC3339, t.String())
	return v
}
