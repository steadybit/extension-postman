// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package utils

import (
	"bytes"
	"encoding/base64"
	"github.com/mitchellh/mapstructure"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/exthttp"
	"io"
	"net/http"
	"os"
)

func WriteActionState[T any](w http.ResponseWriter, state T) {
	err, encodedState := EncodeActionState(state)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to encode action state", err))
	} else {
		exthttp.WriteBody(w, action_kit_api.StatusResult{
			State: &encodedState,
		})
	}
}

func EncodeActionState[T any](attackState T) (error, action_kit_api.ActionState) {
	var result action_kit_api.ActionState
	err := mapstructure.Decode(attackState, &result)
	return err, result
}

func DecodeActionState[T any](attackState action_kit_api.ActionState, result *T) error {
	return mapstructure.Decode(attackState, result)
}

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
