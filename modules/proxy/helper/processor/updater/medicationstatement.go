package updater

import (
	"fmt"

	"github.com/nersus15/mini-proxy/mod-proxy/helper/types"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

func UpdateReferenceMedicationStatement(entry *types.BundleEntry, register *types.SetReference, newPost []types.NewPost) (bool, error) {
	resource, ok := entry.Base.ResourceReal.(fhir.MedicationStatement)
	if !ok {
		return false, fmt.Errorf("failed to cast resource to MedicationStatement")
	}
	for _, baru := range newPost {
		switch baru.ResourceType {
		case "Medication":
			types.UpdateReferenceID(&resource.MedicationReference, register, baru, entry)
		case "Encounter", "EpisodeOfCare":
			types.UpdateReferenceID(resource.Context, register, baru, entry)
		case "Patient":
			types.UpdateReferenceID(&resource.Subject, register, baru, entry)
		}
	}

	resource.Id = entry.Base.Id
	entry.Base.ResourceReal = resource
	return true, nil
}
