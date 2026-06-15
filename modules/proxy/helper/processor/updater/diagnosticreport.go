package updater

import (
	"fmt"

	"github.com/nersus15/mini-proxy/mod-proxy/helper/types"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

func UpdateReferenceDiagnosticReport(entry *types.BundleEntry, register *types.SetReference, newPost []types.NewPost) (bool, error) {
	resource, ok := entry.Base.ResourceReal.(fhir.DiagnosticReport)
	if !ok {
		return false, fmt.Errorf("failed to cast resource to DiagnosticReport")
	}
	for _, baru := range newPost {
		switch baru.ResourceType {
		case "Encounter":
			types.UpdateReferenceID(resource.Encounter, register, baru, entry)
		case "Specimen":
			types.UpdateReferenceArrayID(&resource.Specimen, register, baru, entry)
		case "Observation":
			types.UpdateReferenceArrayID(&resource.Result, register, baru, entry)
		case "ImagingStudy":
			types.UpdateReferenceArrayID(&resource.ImagingStudy, register, baru, entry)
		case "Patient":
			types.UpdateReferenceID(resource.Subject, register, baru, entry)
		case "ServiceRequest", "CarePlan", "MedicationRequest", "NutritionOrder":
			types.UpdateReferenceArrayID(&resource.BasedOn, register, baru, entry)
		}
	}

	resource.Id = entry.Base.Id
	entry.Base.ResourceReal = resource
	return true, nil
}
