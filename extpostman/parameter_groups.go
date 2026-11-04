// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 Steadybit GmbH

package extpostman

import "github.com/steadybit/action-kit/go/action_kit_api/v2"

const (
	groupEnvironment = "Postman Environment"
	groupRunSettings = "Run Settings"
)

var (
	environmentGroup = action_kit_api.ParameterGroup{
		Name:        groupEnvironment,
		Description: "Leave unchanged to run the collection without an environment.",
	}
	runSettingsGroup = action_kit_api.ParameterGroup{
		Name:        groupRunSettings,
		Description: "Leave unchanged to run without timeouts and to continue after failed tests.",
	}
)
