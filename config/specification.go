// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package config

type Specification struct {
	PostmanBaseUrl string `json:"postmanBaseUrl" split_words:"true" required:"false" default:"https://api.getpostman.com"`
}
