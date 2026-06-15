package updater

import (
	"fmt"

	"github.com/nersus15/mini-proxy/mod-proxy/helper/types"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

func UpdateReferenceSpecimen(entry *types.BundleEntry, register *types.SetReference, newPost []types.NewPost) (bool, error) {
	resource, ok := entry.Base.ResourceReal.(fhir.Specimen)
	if !ok {
		return false, fmt.Errorf("failed to cast resource to Specimen")
	}
	for _, baru := range newPost {
		switch baru.ResourceType {
		case "ServiceRequest":
			types.UpdateReferenceArrayID(&resource.Request, register, baru, entry)
		case "Patient":
			types.UpdateReferenceID(resource.Subject, register, baru, entry)
		case "Specimen":
			types.UpdateReferenceArrayID(&resource.Parent, register, baru, entry)
		case "Substance":
			for i := 0; i < len(resource.Container); i++ {
				types.UpdateReferenceID(resource.Container[i].AdditiveReference, register, baru, entry)
			}
			for i := 0; i < len(resource.Processing); i++ {
				types.UpdateReferenceArrayID(&resource.Processing[i].Additive, register, baru, entry)
			}
		}
	}

	resource.Id = entry.Base.Id
	entry.Base.ResourceReal = resource
	return true, nil
}
