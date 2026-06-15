package updater

import (
	"fmt"

	"github.com/nersus15/mini-proxy/mod-proxy/helper/types"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

func UpdateReferenceRelatedPerson(entry *types.BundleEntry, register *types.SetReference, newPost []types.NewPost) (bool, error) {
	resource, ok := entry.Base.ResourceReal.(fhir.RelatedPerson)
	if !ok {
		return false, fmt.Errorf("failed to cast resource to RelatedPerson")
	}
	for _, baru := range newPost {
		switch baru.ResourceType {
		case "Patient":
			types.UpdateReferenceID(&resource.Patient, register, baru, entry)
		}
	}

	resource.Id = entry.Base.Id
	entry.Base.ResourceReal = resource
	return true, nil
}
