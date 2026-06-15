package updater

import (
	"fmt"

	"github.com/nersus15/mini-proxy/mod-proxy/helper/types"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

func UpdateReferenceCarePlan(entry *types.BundleEntry, register *types.SetReference, newPost []types.NewPost) (bool, error) {
	resource, ok := entry.Base.ResourceReal.(fhir.CarePlan)
	if !ok {
		return false, fmt.Errorf("failed to cast resource to CarePlan")
	}
	for _, baru := range newPost {
		switch baru.ResourceType {
		case "Encounter":
			types.UpdateReferenceID(resource.Encounter, register, baru, entry)
		case "CarePlan":
			types.UpdateReferenceArrayID(&resource.BasedOn, register, baru, entry)
		case "Goal":
			types.UpdateReferenceArrayID(&resource.Goal, register, baru, entry)
		case "Patient":
			types.UpdateReferenceID(&resource.Subject, register, baru, entry)
		}
	}

	resource.Id = entry.Base.Id
	entry.Base.ResourceReal = resource
	return true, nil
}
