package updater

import (
	"fmt"

	"github.com/nersus15/mini-proxy/mod-proxy/helper/types"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

func UpdateReferenceProcedure(entry *types.BundleEntry, register *types.SetReference, newPost []types.NewPost) (bool, error) {
	resource, ok := entry.Base.ResourceReal.(fhir.Procedure)
	if !ok {
		return false, fmt.Errorf("failed to cast resource to Procedure")
	}
	for _, baru := range newPost {
		switch baru.ResourceType {
		case "Encounter":
			types.UpdateReferenceID(resource.Encounter, register, baru, entry)
		case "Patient":
			types.UpdateReferenceID(&resource.Subject, register, baru, entry)
		case "CarePlan", "ServiceRequest":
			types.UpdateReferenceArrayID(&resource.BasedOn, register, baru, entry)
		case "Procedure", "Observation":
			types.UpdateReferenceArrayID(&resource.PartOf, register, baru, entry)
			types.UpdateReferenceArrayID(&resource.ReasonReference, register, baru, entry)
		case "Condition", "DiagnosticReport", "DocumentReference":
			types.UpdateReferenceArrayID(&resource.ReasonReference, register, baru, entry)
		case "MedicationAdministration":
			types.UpdateReferenceArrayID(&resource.PartOf, register, baru, entry)
		}
	}

	resource.Id = entry.Base.Id
	entry.Base.ResourceReal = resource
	return true, nil
}
