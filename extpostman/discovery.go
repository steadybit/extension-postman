/*
 * Copyright 2023 steadybit GmbH. All rights reserved.
 */

package extpostman

import (
	"context"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/discovery-kit/go/discovery_kit_commons"
	"github.com/steadybit/discovery-kit/go/discovery_kit_sdk"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/extutil"
	"time"
)

type collectionDiscovery struct {
}

var (
	_ discovery_kit_sdk.TargetDescriber    = (*collectionDiscovery)(nil)
	_ discovery_kit_sdk.AttributeDescriber = (*collectionDiscovery)(nil)
)

func NewPostmanCollectionDiscovery() discovery_kit_sdk.TargetDiscovery {
	discovery := &collectionDiscovery{}
	return discovery_kit_sdk.NewCachedTargetDiscovery(discovery,
		discovery_kit_sdk.WithRefreshTargetsNow(),
		discovery_kit_sdk.WithRefreshTargetsInterval(context.Background(), 1*time.Minute),
	)
}

func (d *collectionDiscovery) Describe() discovery_kit_api.DiscoveryDescription {
	return discovery_kit_api.DiscoveryDescription{
		Id:         targetID,
		RestrictTo: extutil.Ptr(discovery_kit_api.LEADER),
		Discover: discovery_kit_api.DescribingEndpointReferenceWithCallInterval{
			CallInterval: extutil.Ptr("1m"),
		},
	}
}

func (d *collectionDiscovery) DescribeTarget() discovery_kit_api.TargetDescription {
	return discovery_kit_api.TargetDescription{
		Id:      targetID,
		Version: extbuild.GetSemverVersionStringOrUnknown(),
		Icon:    extutil.Ptr(icon),

		// Labels used in the UI
		Label: discovery_kit_api.PluralLabel{One: "Collection", Other: "Collections"},

		// Category for the targets to appear in
		Category: extutil.Ptr("postman"),

		// Specify attributes shown in table columns and to be used for sorting
		Table: discovery_kit_api.Table{
			Columns: []discovery_kit_api.Column{
				{Attribute: "steadybit.label"},
				{Attribute: "postman.collection.id"},
			},
			OrderBy: []discovery_kit_api.OrderBy{
				{
					Attribute: "steadybit.label",
					Direction: "ASC",
				},
			},
		},
	}
}

func (d *collectionDiscovery) DescribeAttributes() []discovery_kit_api.AttributeDescription {
	return []discovery_kit_api.AttributeDescription{
		{
			Attribute: "postman.collection.id",
			Label: discovery_kit_api.PluralLabel{
				One:   "Collection Id",
				Other: "Collection Ids",
			},
		},
	}
}

func (d *collectionDiscovery) DiscoverTargets(_ context.Context) ([]discovery_kit_api.Target, error) {
	collections := GetPostmanCollections()
	targets := make([]discovery_kit_api.Target, len(collections))
	for i, collection := range collections {
		targets[i] = discovery_kit_api.Target{
			Id:         collection.Id,
			TargetType: targetID,
			Label:      collection.Name,
			Attributes: map[string][]string{
				"steadybit.label":     {collection.Name},
				"postman.collection.id":     {collection.Id},
				"postman.collection.name":     {collection.Name},
			},
		}
	}
	return discovery_kit_commons.ApplyAttributeExcludes(targets, []string{}), nil
}
