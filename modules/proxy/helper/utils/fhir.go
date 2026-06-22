package utils

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/nersus15/mini-proxy/mod-proxy/helper/processor"
	"github.com/nersus15/mini-proxy/mod-proxy/helper/types"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	"github.com/webcore-go/webcore/infra/logger"
)

func GetFHIRPatientReference(resource *types.BaseResource) *string {
	if resource == nil || resource.ResourceReal == nil {
		return nil
	}

	val := reflect.ValueOf(resource.ResourceReal)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldName := strings.ToLower(field.Name)
		if fieldName != "subject" && fieldName != "patient" {
			continue
		}

		fieldValue := val.Field(i)
		if fieldValue.Kind() == reflect.Ptr {
			if fieldValue.IsNil() {
				continue
			}
			fieldValue = fieldValue.Elem()
		}

		if fieldValue.Type() != reflect.TypeOf(fhir.Reference{}) {
			continue
		}

		refField := fieldValue.FieldByName("Reference")
		if !refField.IsValid() || refField.Kind() != reflect.Ptr || refField.IsNil() {
			continue
		}

		refStr := refField.Interface().(*string)
		if refStr == nil {
			continue
		}

		patientID := strings.ReplaceAll(*refStr, "Patient/", "")
		return &patientID
	}
	return nil
}

func GetFHIRPatientReferenceFromBundleEntry(entries []fhir.BundleEntry) *string {
	for _, entry := range entries {
		entryBytes, err := json.Marshal(entry.Resource)
		if err != nil {
			continue
		}

		baseResource, err := processor.UnmarshalResource(entryBytes, nil)
		if err != nil {
			logger.Error("Gagal Unmarshal Entry", err)
			continue
		}

		if patientID := GetFHIRPatientReference(baseResource); patientID != nil {
			return patientID
		}
	}
	return nil
}
