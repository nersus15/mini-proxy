package updater

import (
	"fmt"

	"github.com/nersus15/mini-proxy/mod-proxy/helper/types"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

func UpdateReferenceEncounter(entry *types.BundleEntry, register *types.SetReference, newPost []types.NewPost) (bool, error) {
	resource, ok := entry.Base.ResourceReal.(fhir.Encounter)
	if !ok {
		return false, fmt.Errorf("failed to cast resource to Encounter")
	}
	for _, baru := range newPost {
		switch baru.ResourceType {
		case "Condition":
			for i := range resource.Diagnosis {
				types.UpdateReferenceID(&resource.Diagnosis[i].Condition, register, baru, entry)
			}
		case "ServiceRequest":
			types.UpdateReferenceArrayID(&resource.BasedOn, register, baru, entry)
		case "EpisodeOfCare":
			types.UpdateReferenceArrayID(&resource.EpisodeOfCare, register, baru, entry)
		case "Patient":
			types.UpdateReferenceID(resource.Subject, register, baru, entry)
		}
	}

	resource.Id = entry.Base.Id
	entry.Base.ResourceReal = resource
	return true, nil
}
