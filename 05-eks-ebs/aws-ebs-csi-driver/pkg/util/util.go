/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"fmt"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	csi "github.com/container-storage-interface/spec/lib/go/csi"
)

const (
	GiB = int64(1024 * 1024 * 1024)
)

var (
	isAlphanumericRegex = regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString
)

// RoundUpBytes rounds up the volume size in bytes up to multiplications of GiB
func RoundUpBytes(volumeSizeBytes int64) int64 {
	return roundUpSize(volumeSizeBytes, GiB) * GiB
}

// RoundUpGiB rounds up the volume size in bytes upto multiplications of GiB
// in the unit of GiB
func RoundUpGiB(volumeSizeBytes int64) (int32, error) {
	result := roundUpSize(volumeSizeBytes, GiB)
	if result > int64(math.MaxInt32) {
		return 0, fmt.Errorf("rounded up size exceeds maximum value of int32: %d", result)
	}
	return int32(result), nil
}

// BytesToGiB converts Bytes to GiB
func BytesToGiB(volumeSizeBytes int64) int32 {
	result := volumeSizeBytes / GiB
	if result > int64(math.MaxInt32) {
		// Handle overflow
		return math.MaxInt32
	}
	return int32(result)
}

// GiBToBytes converts GiB to Bytes
func GiBToBytes(volumeSizeGiB int32) int64 {
	return int64(volumeSizeGiB) * GiB
}

func ParseEndpoint(endpoint string) (string, string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", "", fmt.Errorf("could not parse endpoint: %w", err)
	}

	addr := filepath.Join(u.Host, filepath.FromSlash(u.Path))

	scheme := strings.ToLower(u.Scheme)
	switch scheme {
	case "tcp":
	case "unix":
		addr = filepath.Join("/", addr)
		if err := os.Remove(addr); err != nil && !os.IsNotExist(err) {
			return "", "", fmt.Errorf("could not remove unix domain socket %q: %w", addr, err)
		}
	default:
		return "", "", fmt.Errorf("unsupported protocol: %s", scheme)
	}

	return scheme, addr, nil
}

func roundUpSize(volumeSizeBytes int64, allocationUnitBytes int64) int64 {
	if allocationUnitBytes == 0 {
		return 0 // Avoid division by zero
	}
	return (volumeSizeBytes + allocationUnitBytes - 1) / allocationUnitBytes
}

// GetAccessModes returns a slice containing all of the access modes defined
// in the passed in VolumeCapabilities.
func GetAccessModes(caps []*csi.VolumeCapability) *[]string {
	modes := []string{}
	for _, c := range caps {
		modes = append(modes, c.GetAccessMode().GetMode().String())
	}
	return &modes
}

func IsSBE(region string) bool {
	return region == "snow"
}

// StringIsAlphanumeric returns true if a given string contains only English letters or numbers
func StringIsAlphanumeric(s string) bool {
	return isAlphanumericRegex(s)
}
