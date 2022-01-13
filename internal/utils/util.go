package utils

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

var idMutex sync.Mutex
var idCounter uint32

func PrefixedUniqueId(prefix string) string {
	// Be precise to 4 digits of fractional seconds, but remove the dot before the
	// fractional seconds.
	timestamp := strings.Replace(
		time.Now().UTC().Format("20060102150405.0000"), ".", "", 1)

	idMutex.Lock()
	defer idMutex.Unlock()
	idCounter++
	return fmt.Sprintf("%s%s%08x", prefix, timestamp, idCounter)
}

func DifferenceSlice(slice1, slice2 []string) []string {
	m := make(map[string]int)
	nn := make([]string, 0)
	inter := IntersectSlice(slice1, slice2)
	for _, v := range inter {
		m[v]++
	}

	for _, value := range slice1 {
		times, _ := m[value]
		if times == 0 {
			nn = append(nn, value)
		}
	}
	return nn
}

func IntersectSlice(slice1, slice2 []string) []string {
	m := make(map[string]int)
	nn := make([]string, 0)
	for _, v := range slice1 {
		m[v]++
	}

	for _, v := range slice2 {
		times, _ := m[v]
		if times == 1 {
			nn = append(nn, v)
		}
	}
	return nn
}

func CheckError(err error) {
	if err != nil {
		fmt.Println("[ERROR]", err)
	}
}

func CheckErrorWithCode(err error) {
	if err != nil {
		fmt.Println("[ERROR]", err)
		os.Exit(1)
	}
}

func MustExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}
