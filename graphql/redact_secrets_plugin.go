package graphql

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"

	"github.com/99designs/gqlgen/codegen/config"
	"github.com/evergreen-ci/evergreen/util"
	"github.com/mongodb/grip"
	"github.com/mongodb/grip/message"
)

// GenerateSecretFields generates a file that contains a list of fields that should be redacted in logs.
func GenerateSecretFields(cfg *config.Config) error {
	redactedFields := make(map[string]bool)

	for _, schemaType := range cfg.Schema.Types {
		for _, field := range schemaType.Fields {
			if field.Directives.ForName("redactSecrets") != nil {
				redactedFields[field.Name] = true
			}
		}
	}

	var fields []string
	for field := range redactedFields {
		fields = append(fields, field)
	}

	// Sort the fields to ensure consistent output.
	sort.Strings(fields)

	return generateRedactedFieldsFile(fields)
}

// generateRedactedFieldsFile generates a file that contains a list of fields that should be redacted in logs. It is used to generate the redacted_fields_gen.go file.
// The file contains a list of fields that should be redacted in logs, a function to check if a field should be redacted, and a function to redact fields in a map.
func generateRedactedFieldsFile(fields []string) error {
	file, err := os.Create("graphql/redacted_fields_gen.go")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(`// Code generated by graphql/redact_secrets_plugin.go DO NOT EDIT.
package graphql

var redactedFields = map[string]bool{
`)
	if err != nil {
		return err
	}

	for _, field := range fields {
		_, err = file.WriteString(fmt.Sprintf("\t\"%s\": true,\n", field))
		if err != nil {
			return err
		}
	}

	_, err = file.WriteString("}\n")
	if err != nil {
		return err
	}
	return nil
}

func isFieldRedacted(fieldName string, fieldsToRedact map[string]bool) bool {
	_, ok := fieldsToRedact[fieldName]
	return ok
}

// RedactFieldsInMap recursively searches for and redacts fields in a map.
// Assumes map structure like map[string]interface{} where interface{} can be another map, a slice, or a basic datatype.
func RedactFieldsInMap(data map[string]interface{}, fieldsToRedact map[string]bool) map[string]interface{} {
	dataCopy := map[string]interface{}{}
	registeredTypes := []interface{}{
		map[interface{}]interface{}{},
		map[string]interface{}{},
		[]interface{}{},
		[]util.KeyValuePair{},
		json.Number(""),
	}
	if err := util.DeepCopy(data, &dataCopy, registeredTypes); err != nil {
		// If theres an error copying the data, log it and return an empty map.
		grip.Error(message.WrapError(err, message.Fields{
			"message": "failed to deep copy request variables",
		}))
		return map[string]interface{}{}
	}
	recursivelyRedactFieldsInMap(dataCopy, fieldsToRedact)

	return dataCopy
}

func recursivelyRedactFieldsInMap(data map[string]interface{}, fieldsToRedact map[string]bool) {
	for key, value := range data {
		// If the current key matches a field that should be redacted, redact it.
		if isFieldRedacted(key, fieldsToRedact) {
			data[key] = "REDACTED"
			continue
		}

		// Handle nil values.
		if value == nil {
			continue
		}
		// If the value is a map, recursively redact fields within it.
		if reflect.TypeOf(value).Kind() == reflect.Map {
			if subMap, ok := value.(map[string]interface{}); ok {
				recursivelyRedactFieldsInMap(subMap, fieldsToRedact)
			}
		}

		// If the value is a slice, iterate over it and redact fields if elements are maps.
		if reflect.TypeOf(value).Kind() == reflect.Slice {
			sliceVal := reflect.ValueOf(value)
			for i := 0; i < sliceVal.Len(); i++ {
				elem := sliceVal.Index(i).Interface()
				if reflect.TypeOf(elem).Kind() == reflect.Map {
					if elemMap, ok := elem.(map[string]interface{}); ok {
						recursivelyRedactFieldsInMap(elemMap, fieldsToRedact)
					}
				}
			}
		}
	}
}
