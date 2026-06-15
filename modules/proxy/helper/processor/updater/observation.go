package updater

import (
	"fmt"

	"github.com/nersus15/mini-proxy/mod-proxy/helper/types"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	"github.com/webcore-go/webcore/infra/logger"
)

func UpdateReferenceObservation(entry *types.BundleEntry, register *types.SetReference, newPost []types.NewPost) (bool, error) {
	resource, ok := entry.Base.ResourceReal.(fhir.Observation)
	if !ok {
		return false, fmt.Errorf("failed to cast resource to Observation")
	}
	logger.InfoJson("Observation", resource)
	for _, baru := range newPost {
		switch baru.ResourceType {
		case "Encounter":
			types.UpdateReferenceID(resource.Encounter, register, baru, entry)
		case "Specimen":
			types.UpdateReferenceID(resource.Specimen, register, baru, entry)
		case "ServiceRequest", "CarePlan", "MedicationRequest", "NutritionOrder":
			types.UpdateReferenceArrayID(&resource.BasedOn, register, baru, entry)
		case "Patient":
			types.UpdateReferenceID(resource.Subject, register, baru, entry)
		case "RelatedPerson":
			types.UpdateReferenceArrayID(&resource.Performer, register, baru, entry)

		}
	}

	resource.Id = entry.Base.Id
	entry.Base.ResourceReal = resource
	return true, nil
}
