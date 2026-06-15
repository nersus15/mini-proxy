package updater

import (
	"fmt"

	"github.com/nersus15/mini-proxy/mod-proxy/helper/types"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

func UpdateReferencePractitionerRole(entry *types.BundleEntry, register *types.SetReference, newPost []types.NewPost) (bool, error) {
	resource, ok := entry.Base.ResourceReal.(fhir.PractitionerRole)
	if !ok {
		return false, fmt.Errorf("failed to cast resource to PractitionerRole")
	}
	for _, baru := range newPost {
		switch baru.ResourceType {
		case "Practitioner":
			types.UpdateReferenceID(resource.Practitioner, register, baru, entry)
		}
	}

	resource.Id = entry.Base.Id
	entry.Base.ResourceReal = resource
	return true, nil
}
