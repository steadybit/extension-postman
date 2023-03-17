// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package main

import (
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extlogging"
	"github.com/steadybit/extension-postman/extpostman"
)

func main() {
	extlogging.InitZeroLog()
	extbuild.PrintBuildInformation()

	exthttp.RegisterHttpHandler("/", exthttp.GetterAsHandler(getActionList))

	extpostman.RegisterHandlers()

	exthttp.Listen(exthttp.ListenOpts{
		Port: 8086,
	})
}

func getActionList() action_kit_api.ActionList {
	return action_kit_api.ActionList{
		Actions: []action_kit_api.DescribingEndpointReference{
			{
				"GET",
				"/postman/collection/run",
			},
		},
	}
}
