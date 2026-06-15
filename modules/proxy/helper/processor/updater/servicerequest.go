package updater

import (
	"fmt"

	"github.com/nersus15/mini-proxy/mod-proxy/helper/types"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

func UpdateReferenceServiceRequest(entry *types.BundleEntry, register *types.SetReference, newPost []types.NewPost) (bool, error) {
	resource, ok := entry.Base.ResourceReal.(fhir.ServiceRequest)
	if !ok {
		return false, fmt.Errorf("failed to cast resource to ServiceRequest")
	}
	for _, baru := range newPost {
		switch baru.ResourceType {
		case "Encounter":
			types.UpdateReferenceID(resource.Encounter, register, baru, entry)
		case "Specimen":
			types.UpdateReferenceArrayID(&resource.Specimen, register, baru, entry)
		case "ServiceRequest":
			types.UpdateReferenceArrayID(&resource.BasedOn, register, baru, entry)
		case "CarePlan":
			types.UpdateReferenceArrayID(&resource.BasedOn, register, baru, entry)
		case "Patient":
			types.UpdateReferenceID(&resource.Subject, register, baru, entry)
			types.UpdateReferenceID(resource.Requester, register, baru, entry)
			types.UpdateReferenceArrayID(&resource.Performer, register, baru, entry)
		case "RelatedPerson":
			types.UpdateReferenceID(resource.Requester, register, baru, entry)
			types.UpdateReferenceArrayID(&resource.Performer, register, baru, entry)
		case "Observation", "DiagnosticReport":
			types.UpdateReferenceArrayID(&resource.SupportingInfo, register, baru, entry)
			types.UpdateReferenceArrayID(&resource.ReasonReference, register, baru, entry)
		case "Condition":
			types.UpdateReferenceArrayID(&resource.ReasonReference, register, baru, entry)
		case "DocumentReference":
			types.UpdateReferenceArrayID(&resource.ReasonReference, register, baru, entry)
		case "CareTeam", "HealthService":
			types.UpdateReferenceArrayID(&resource.Performer, register, baru, entry)
		// case "Practitioner", "Organization", "PractitionerRole":
		// 	types.UpdateReferenceID(resource.Requester, register, baru, entry)
		// 	types.UpdateReferenceArrayID(&resource.Performer, register, baru, entry)
		case "Location":
			types.UpdateReferenceArrayID(&resource.LocationReference, register, baru, entry)
		}
	}

	resource.Id = entry.Base.Id
	entry.Base.ResourceReal = resource
	return true, nil
}
