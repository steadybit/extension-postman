// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package main

import (
	"github.com/rs/zerolog"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/exthealth"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extlogging"
	"github.com/steadybit/extension-kit/extruntime"
	"github.com/steadybit/extension-postman/config"
	"github.com/steadybit/extension-postman/extpostman"
)

func main() {
	config.ParseConfiguration()

	extlogging.InitZeroLog()
	extbuild.PrintBuildInformation()
	extruntime.LogRuntimeInformation(zerolog.DebugLevel)
	exthealth.StartProbes(8087)

	action_kit_sdk.RegisterCoverageEndpoints()
	action_kit_sdk.RegisterAction(extpostman.NewPostmanAction())
	action_kit_sdk.InstallSignalHandler()

	exthttp.RegisterHttpHandler("/", exthttp.GetterAsHandler(action_kit_sdk.GetActionList))
	exthttp.Listen(exthttp.ListenOpts{
		Port: 8086,
	})
}
