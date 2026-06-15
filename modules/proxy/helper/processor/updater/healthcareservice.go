package updater

import (
	"fmt"

	"github.com/nersus15/mini-proxy/mod-proxy/helper/types"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

func UpdateReferenceHealthcareService(entry *types.BundleEntry, register *types.SetReference, newPost []types.NewPost) (bool, error) {
	resource, ok := entry.Base.ResourceReal.(fhir.HealthcareService)
	if !ok {
		return false, fmt.Errorf("failed to cast resource to HealthcareService")
	}
	for _, baru := range newPost {
		switch baru.ResourceType {
		case "Location":
			types.UpdateReferenceArrayID(&resource.Location, register, baru, entry)
		}
	}

	resource.Id = entry.Base.Id
	entry.Base.ResourceReal = resource
	return true, nil
}
