// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 Steadybit GmbH

package extpostman

import (
	"testing"

	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/stretchr/testify/assert"
)

func TestParameterGroups(t *testing.T) {
	desc := NewPostmanAction().Describe()

	var order []string
	members := map[string][]string{}
	for _, p := range desc.Parameters {
		order = append(order, p.Name)
		assert.Nil(t, p.Order, "parameter %s: the list order decides", p.Name)
		if p.Group != nil {
			members[*p.Group] = append(members[*p.Group], p.Name)
		} else {
			// No advanced parameter is left alone in "Additional Settings" on platforms with groups.
			assert.False(t, p.Advanced != nil && *p.Advanced, "parameter %s", p.Name)
		}
	}

	// Iterations is shown right below the estimated duration: a run ends after its iterations, not after the duration.
	assert.Equal(t, []string{"duration", "iterations"}, order[:2])
	assert.Equal(t, map[string][]string{
		groupEnvironment: {"environmentIdOrName", "environment"},
		groupRunSettings: {"timeout", "timeoutRequest", "verbose", "bail"},
	}, members)
	assert.Equal(t, []action_kit_api.ParameterGroup{environmentGroup, runSettingsGroup}, *desc.ParameterGroups)
}
