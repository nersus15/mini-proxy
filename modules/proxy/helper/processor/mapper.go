package processor

import (
	"github.com/goccy/go-json"
	"github.com/nersus15/mini-proxy/mod-proxy/helper/processor/updater"

	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

var _ProcessorMapFunction = map[string]ProcessorMapEntry{
	"Encounter": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalEncounter(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			Encounter := Resource.(fhir.Encounter)
			return Encounter.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceEncounter,
	},
	"Condition": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalCondition(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			Condition := Resource.(fhir.Condition)
			return Condition.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceCondition,
	},
	"Observation": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalObservation(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			Observation := Resource.(fhir.Observation)
			return Observation.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceObservation,
	},
	"ServiceRequest": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalServiceRequest(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			ServiceRequest := Resource.(fhir.ServiceRequest)
			return ServiceRequest.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceServiceRequest,
	},
	"Specimen": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalSpecimen(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			Specimen := Resource.(fhir.Specimen)
			return Specimen.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceSpecimen,
	},
	"DiagnosticReport": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalDiagnosticReport(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			DiagnosticReport := Resource.(fhir.DiagnosticReport)
			return DiagnosticReport.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceDiagnosticReport,
	},
	"Medication": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalMedication(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			Medication := Resource.(fhir.Medication)
			return Medication.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceMedication,
	},
	"MedicationRequest": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalMedicationRequest(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			MedicationRequest := Resource.(fhir.MedicationRequest)
			return MedicationRequest.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceMedicationRequest,
	},
	"MedicationStatement": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalMedicationStatement(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			MedicationStatement := Resource.(fhir.MedicationStatement)
			return MedicationStatement.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceMedicationStatement,
	},
	"MedicationDispense": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalMedicationDispense(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			MedicationDispense := Resource.(fhir.MedicationDispense)
			return MedicationDispense.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceMedicationDispense,
	},
	"MedicationAdministration": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalMedicationAdministration(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			MedicationAdministration := Resource.(fhir.MedicationAdministration)
			return MedicationAdministration.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceMedicationAdministration,
	},
	"Procedure": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalProcedure(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			Procedure := Resource.(fhir.Procedure)
			return Procedure.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceProcedure,
	},
	"ImagingStudy": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalImagingStudy(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			ImagingStudy := Resource.(fhir.ImagingStudy)
			return ImagingStudy.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceImagingStudy,
	},
	"QuestionnaireResponse": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalQuestionnaireResponse(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			QuestionnaireResponse := Resource.(fhir.QuestionnaireResponse)
			return QuestionnaireResponse.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceQuestionnaireResponse,
	},
	"Composition": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalComposition(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			Composition := Resource.(fhir.Composition)
			return Composition.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceComposition,
	},
	"AllergyIntolerance": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalAllergyIntolerance(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			AllergyIntolerance := Resource.(fhir.AllergyIntolerance)
			return AllergyIntolerance.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceAllergyIntolerance,
	},
	"CarePlan": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalCarePlan(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			CarePlan := Resource.(fhir.CarePlan)
			return CarePlan.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceCarePlan,
	},
	"ClinicalImpression": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalClinicalImpression(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			ClinicalImpression := Resource.(fhir.ClinicalImpression)
			return ClinicalImpression.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceClinicalImpression,
	},
	"FamilyMemberHistory": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalFamilyMemberHistory(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			FamilyMemberHistory := Resource.(fhir.FamilyMemberHistory)
			return FamilyMemberHistory.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceFamilyMemberHistory,
	},
	"Immunization": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalImmunization(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			Immunization := Resource.(fhir.Immunization)
			return Immunization.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceImmunization,
	},
	"EpisodeOfCare": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalEpisodeOfCare(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			EpisodeOfCare := Resource.(fhir.EpisodeOfCare)
			return EpisodeOfCare.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceEpisodeOfCare,
	},
	"NutritionOrder": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalNutritionOrder(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			NutritionOrder := Resource.(fhir.NutritionOrder)
			return NutritionOrder.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceNutritionOrder,
	},
	"Organization": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalOrganization(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			Organization := Resource.(fhir.Organization)
			return Organization.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceOrganization,
	},
	"Patient": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalPatient(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			Patient := Resource.(fhir.Patient)
			return Patient.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferencePatient,
	},
	"RelatedPerson": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalRelatedPerson(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			RelatedPerson := Resource.(fhir.RelatedPerson)
			return RelatedPerson.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceRelatedPerson,
	},
	"Practitioner": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalPractitioner(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			Practitioner := Resource.(fhir.Practitioner)
			return Practitioner.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferencePractitioner,
	},
	"HealthcareService": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalHealthcareService(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			HealthcareService := Resource.(fhir.HealthcareService)
			return HealthcareService.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceHealthcareService,
	},
	"Location": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalLocation(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			Location := Resource.(fhir.Location)
			return Location.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceLocation,
	},
	"Substance": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalSubstance(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			Substance := Resource.(fhir.Substance)
			return Substance.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceSubstance,
	},
	"RiskAssessment": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalRiskAssessment(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			RiskAssessment := Resource.(fhir.RiskAssessment)
			return RiskAssessment.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceRiskAssessment,
	},
	"Goal": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalGoal(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			Goal := Resource.(fhir.Goal)
			return Goal.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceGoal,
	},
	"PractitionerRole": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalPractitionerRole(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			PractitionerRole := Resource.(fhir.PractitionerRole)
			return PractitionerRole.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferencePractitionerRole,
	},
	"Bundle": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalBundle(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			Bundle := Resource.(fhir.Bundle)
			return Bundle.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceOperationOutcome,
	},
	"OperationOutcome": {
		Unmarshaller: func(Raw json.RawMessage, id *string) (any, error) {
			resource, err := fhir.UnmarshalOperationOutcome(Raw)
			if err == nil && id != nil {
				resource.Id = id
			}
			return resource, err
		},
		Marshaller: func(Resource any) ([]byte, error) {
			OperationOutcome := Resource.(fhir.OperationOutcome)
			return OperationOutcome.MarshalJSON()
		},
		ReferenceUpdater: updater.UpdateReferenceOperationOutcome,
	},
}
