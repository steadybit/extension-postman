// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package extension_kit

import "github.com/steadybit/extension-kit/extutil"

type ExtensionError struct {
	// A human-readable explanation specific to this occurrence of the problem.
	Detail *string `json:"detail,omitempty"`

	// A URI reference that identifies the specific occurrence of the problem.
	Instance *string `json:"instance,omitempty"`

	// A short, human-readable summary of the problem type.
	Title string `json:"title"`

	// A URI reference that identifies the problem type.
	Type *string `json:"type,omitempty"`
}

func ToError(title string, err error) ExtensionError {
	var response ExtensionError
	if err != nil {
		response = ExtensionError{Title: title, Detail: extutil.Ptr(err.Error())}
	} else {
		response = ExtensionError{Title: title}
	}
	return response
}
