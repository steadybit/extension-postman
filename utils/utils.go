// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package utils

import (
	"bytes"
	"encoding/base64"
	"github.com/mitchellh/mapstructure"
	"github.com/steadybit/attack-kit/go/attack_kit_api"
	"github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/exthttp"
	"io"
	"net/http"
	"os"
)

func WriteAttackState[T any](w http.ResponseWriter, state T) {
	err, encodedState := EncodeAttackState(state)
	if err != nil {
		exthttp.WriteError(w, extension_kit.ToError("Failed to encode attack state", err))
	} else {
		exthttp.WriteBody(w, attack_kit_api.StatusResult{
			State: &encodedState,
		})
	}
}

func EncodeAttackState[T any](attackState T) (error, attack_kit_api.AttackState) {
	var result attack_kit_api.AttackState
	err := mapstructure.Decode(attackState, &result)
	return err, result
}

func DecodeAttackState[T any](attackState attack_kit_api.AttackState, result *T) error {
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
