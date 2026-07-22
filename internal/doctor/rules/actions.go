package rules

import "github.com/baselhusam/bareai-cli/internal/snapshot"

func containerOffers(containerID, containerName string, verbs ...string) []snapshot.ActionOffer {
	ref := containerName
	if ref == "" {
		ref = containerID
	}
	if ref == "" {
		return nil
	}
	offers := make([]snapshot.ActionOffer, 0, len(verbs))
	for _, verb := range verbs {
		offers = append(offers, snapshot.ActionOffer{
			Verb:       verb,
			TargetKind: "container",
			TargetRef:  ref,
			Summary:    ref,
		})
	}
	return offers
}

func endpointOffer(endpoint, verb string) []snapshot.ActionOffer {
	if endpoint == "" {
		return nil
	}
	return []snapshot.ActionOffer{{
		Verb:       verb,
		TargetKind: "endpoint",
		TargetRef:  endpoint,
		Summary:    endpoint,
	}}
}
