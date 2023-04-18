// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package main

import (
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/exthealth"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extlogging"
	"github.com/steadybit/extension-postman/extpostman"
)

func main() {
	extlogging.InitZeroLog()
	extbuild.PrintBuildInformation()
	exthealth.StartProbes(8087)

	action_kit_sdk.RegisterAction(extpostman.NewPostmanAction())
	action_kit_sdk.InstallSignalHandler()

	exthttp.RegisterHttpHandler("/", exthttp.GetterAsHandler(action_kit_sdk.GetActionList))
	exthttp.Listen(exthttp.ListenOpts{
		Port: 8086,
	})
}
