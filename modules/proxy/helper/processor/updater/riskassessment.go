package updater

import (
	"fmt"

	"github.com/nersus15/mini-proxy/mod-proxy/helper/types"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

func UpdateReferenceRiskAssessment(entry *types.BundleEntry, register *types.SetReference, newPost []types.NewPost) (bool, error) {
	resource, ok := entry.Base.ResourceReal.(fhir.RiskAssessment)
	if !ok {
		return false, fmt.Errorf("failed to cast resource to RiskAssessment")
	}
	for _, baru := range newPost {
		switch baru.ResourceType {
		case "Encounter":
			types.UpdateReferenceID(resource.Encounter, register, baru, entry)
		case "Condition":
			types.UpdateReferenceID(resource.Condition, register, baru, entry)
			types.UpdateReferenceArrayID(&resource.ReasonReference, register, baru, entry)
		case "Patient":
			types.UpdateReferenceID(&resource.Subject, register, baru, entry)
		}
	}

	resource.Id = entry.Base.Id
	entry.Base.ResourceReal = resource
	return true, nil
}
