// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package utils

import (
	"bytes"
	"encoding/base64"
	"io"
	"os"
)

func RemoveAtIndex[T any](s []T, index int) []T {
	return append(s[:index], s[index+1:]...)
}

func File2Base64(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, f)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buffer.Bytes()), nil
}
