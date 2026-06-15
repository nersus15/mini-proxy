package processor

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/nersus15/mini-proxy/mod-proxy/helper/types"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	"github.com/webcore-go/webcore/infra/logger"
)

type ProcessorMapEntry struct {
	Unmarshaller     func(raw json.RawMessage, id *string) (any, error)
	Marshaller       func(resource any) ([]byte, error)
	ReferenceUpdater func(entry *types.BundleEntry, register *types.SetReference, newPost []types.NewPost) (bool, error)
}

func NewBaseResource(raw json.RawMessage, id *string) (*types.BaseResource, error) {
	base, err := UnmarshalResource(raw, id)
	if err != nil {
		return nil, err
	}

	// Resource dengan field 'resourceType' terdefinisi dan ada 'ResourceReal' yang nanti dijadikan acuan Marshalling
	return base, nil
}

func Marshal(base *types.BaseResource) ([]byte, error) {
	processor, ok := _ProcessorMapFunction[*base.ResourceType]
	if ok && processor.Marshaller != nil {
		return processor.Marshaller(base.ResourceReal)
	} else {
		return nil, fmt.Errorf("marshal resourceType %s is not implemented yet", *base.ResourceType)
	}
}

func UnmarshalResource(raw json.RawMessage, id *string) (*types.BaseResource, error) {
	base := types.BaseResource{}
	err1 := json.Unmarshal(raw, &base)
	if err1 != nil {
		return nil, err1
	}

	if base.ResourceType == nil {
		return nil, fmt.Errorf("resourceType is nil")
	}

	var resource any
	var err2 error
	processor, ok := _ProcessorMapFunction[*base.ResourceType]
	if ok && processor.Unmarshaller != nil {
		resource, err2 = processor.Unmarshaller(raw, id)
	} else {
		return nil, fmt.Errorf("unmarshal resourceType %s is not implemented yet", *base.ResourceType)
	}

	if err2 != nil {
		return nil, err2
	}

	if id != nil {
		base.Id = id
	}

	base.ResourceReal = resource

	return &base, nil
}

func UpdateTemporaryReference(entry *types.BundleEntry, register *types.SetReference, idBaru []types.NewPost) bool {
	processor, ok := _ProcessorMapFunction[*entry.Base.ResourceType]
	if ok && processor.ReferenceUpdater != nil {
		ok, _ := processor.ReferenceUpdater(entry, register, idBaru)
		return ok
	} else {
		logger.Info("update reference for resourceType", *entry.Base.ResourceType, "is not implemented yet")
	}

	return false
}

func IsTemporaryId(reference fhir.Reference) bool {
	return reference.Reference == nil || !strings.HasPrefix(*reference.Reference, "urn:uuid:")
}

func ExtractReferences(resource types.BaseResource) []fhir.Reference {
	var references []fhir.Reference

	// Use reflection to find all reference fields in the resource
	val := reflect.ValueOf(resource.ResourceReal)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() == reflect.Struct {
		// logger.Debug("Extract " + val.Type().Name())
		references = findReferencesInStruct(val /*, ""*/)
	}

	return references
}

func findReferencesInStruct(val reflect.Value /*, pad string*/) []fhir.Reference {
	var references []fhir.Reference

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// n := field.Type().Name()
		// if field.Kind() == reflect.Ptr {
		// 	n = "*" + field.Type().Elem().Name()
		// } else if field.Kind() == reflect.Slice {
		// 	n = "[]" + field.Type().Elem().Name()
		// }
		// logger.Debug(pad + "  + " + fieldType.Name + " (" + n + ")")

		// Check if this field is a Reference
		switch field.Kind() {
		case reflect.Ptr:
			if field.IsNil() {
				continue
			}

			if fieldType.Type == reflect.TypeOf(&fhir.Reference{}) {
				// logger.Debug(pad + "     >> DITEMUKAN STRUCT POINTER")
				ref := field.Interface().(*fhir.Reference)
				if ref.Reference != nil && *ref.Reference != "" {
					references = append(references, *ref)
				}
			} else if fieldType.Type == reflect.TypeOf(&[]fhir.Reference{}) {
				// logger.Debug(pad + "     >> DITEMUKAN ARRAY POINTER")
				refs := field.Interface().(*[]fhir.Reference)
				for _, ref := range *refs {
					if ref.Reference != nil && *ref.Reference != "" {
						references = append(references, ref)
					}
				}
			}
		case reflect.Slice:
			if field.IsNil() {
				continue
			}

			if fieldType.Type == reflect.TypeOf([]fhir.Reference{}) {
				// Check if this field is a slice of References
				// logger.Debug(pad + "     >> DITEMUKAN ARRAY")
				refs := field.Interface().([]fhir.Reference)
				for _, ref := range refs {
					if ref.Reference != nil && *ref.Reference != "" {
						references = append(references, ref)
					}
				}
			} else if field.Type().Elem().Kind() == reflect.Struct {
				// Check slices of structs
				for j := 0; j < field.Len(); j++ {
					references = append(references, findReferencesInStruct(field.Index(j) /*, pad+"  "*/)...)
				}
			}
		default:
			if fieldType.Type == reflect.TypeOf(fhir.Reference{}) {
				// logger.Debug(pad + "     >> DITEMUKAN STRUCT")
				ref := field.Interface().(fhir.Reference)
				if ref.Reference != nil && *ref.Reference != "" {
					references = append(references, ref)
				}
			} else if field.Kind() == reflect.Struct {
				// Recursively check nested structs
				references = append(references, findReferencesInStruct(field /*, pad+"  "*/)...)
			}
		}
	}

	return references
}
