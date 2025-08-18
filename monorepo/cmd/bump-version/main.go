package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/mod/semver"
)

func main() {
	if len(os.Args) != 3 {
		panic("Usage: update-version-file path/to/VERSION major|minor|patch")
	}

	path := os.Args[1]
	kind := os.Args[2]

	input, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	version := strings.TrimSpace(string(input))
	version = semver.Canonical(version)
	if version[0] != 'v' {
		panic("invalid version format")
	}
	s := strings.Split(version[1:], ".")
	if len(s) != 3 {
		panic("invalid version format, expected major.minor.patch")
	}
	major, err := strconv.Atoi(s[0])
	if err != nil {
		panic(err)
	}
	minor, err := strconv.Atoi(s[1])
	if err != nil {
		panic(err)
	}
	patch, err := strconv.Atoi(s[2])
	if err != nil {
		panic(err)
	}

	switch kind {
	case "major":
		major++
	case "minor":
		minor++
	case "patch":
		patch++
	default:
		panic("unknown version kind")
	}

	newVersion := fmt.Sprintf("v%d.%d.%d", major, minor, patch)
	fmt.Printf("Updating version from %s to %s\n", version, newVersion)
	if err := os.WriteFile(path, []byte(newVersion), 0644); err != nil {
		panic(err)
	}
}
