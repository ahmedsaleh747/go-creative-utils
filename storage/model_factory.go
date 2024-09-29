package storage

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"
)

type Model interface {
}

var modelConfig = map[string]map[string]any{}
var typeRegistry = map[string]func() interface{}{}

func AddConfig(model interface{}) {
	title := callFunctionGeneric(model, "GetTitle")
	apiUrl := callFunctionGeneric(model, "GetApiUrl")
	extractModelConfig(model, title, apiUrl)
}

func extractModelConfig(model interface{}, title string, apiUrl string) {
	modelType := reflect.TypeOf(model).Elem()
	var fields []map[string]interface{}

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)

		//Don't consider the primary key
		fieldZad := field.Tag.Get("zad")
		if strings.Contains(fieldZad, "hidden") {
			continue
		}

		//Consider the case where json tag has commas, just use the first tab value
		fieldName := field.Tag.Get("json")
		if strings.Contains(fieldName, ",") {
			fieldName = strings.Split(fieldName, ",")[0]
		} else if fieldName == "-" {
			fieldName = strings.ToLower(field.Name)
		}

		fieldInfo := map[string]interface{}{
			"name":  fieldName,
			"label": field.Name,
		}
		if strings.Contains(fieldZad, "optional") {
			fieldInfo["optional"] = true
		}
		if strings.Contains(fieldZad, "block") {
			fieldInfo["block"] = true
		}
		if strings.Contains(fieldZad, "chartData") {
			fieldInfo["chartData"] = true
		}
		if strings.Contains(fieldZad, "tags") {
			fieldInfo["tags"] = true
		}
		if strings.Contains(fieldZad, "short-span") {
			fieldInfo["short-span"] = true
		}
		if strings.Contains(fieldZad, "masterSelector") {
			if configValue, ok := getFieldConfigValue(fieldZad, "masterSelector:"); ok {
				fieldInfo["masterSelector"] = configValue
			}
		}
		if strings.Contains(fieldZad, "href") {
			if configValue, ok := getFieldConfigValue(fieldZad, "href:"); ok {
				fieldInfo["href"] = configValue
			}
		}
		if strings.Contains(fieldZad, "enum") {
			if configValue, ok := getFieldConfigValue(fieldZad, "enum:"); ok {
				fieldInfo["type"] = "select"
				fieldInfo["allowedValues"] = strings.Split(configValue, "|")
			}
		}

		fieldGorm := field.Tag.Get("gorm")
		if strings.Contains(fieldGorm, "foreignKey") || strings.Contains(fieldZad, "enum") {
			if configValue, ok := getFieldConfigValue(fieldGorm, "foreignKey:"); ok {
				fieldInfo["name"] = configValue
			}

			fieldInfo["type"] = "select"
			if strings.Contains(fieldZad, "enum") {
				fieldInfo["selectorOf"] = "enum"
			} else {
				fieldInfo["selectorOf"] = field.Type.Name()
			}
			log.Printf("Constructing model %s configuration, field %s is a selector of %s", modelType, field.Name, fieldInfo["selectorOf"])
		} else {
			// Determine the type for JSON schema
			switch field.Type.Kind() {
			case reflect.String:
				if strings.Contains(fieldZad, "sensitive") {
					fieldInfo["type"] = "password"
				} else {
					fieldInfo["type"] = "text"
				}
				log.Printf("Constructing model %s configuration, field %s is a text", modelType, field.Name)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint:
				log.Printf("Constructing model %s configuration, field %s is a number", modelType, field.Name)
				fieldInfo["type"] = "number"
			case reflect.Bool:
				log.Printf("Constructing model %s configuration, field %s is a boolean", modelType, field.Name)
				fieldInfo["type"] = "bool"
			case reflect.Struct:
				if field.Type == reflect.TypeOf(time.Time{}) {
					fieldInfo["type"] = "date"
				}
			default:
			}
		}
		fields = append(fields, fieldInfo)
	}

	var actions = []string{}
	if actionsStr := callFunctionGeneric(model, "ExtraActions"); actionsStr != "" {
		actions = strings.Split(actionsStr, ",")
	}

	log.Printf("Storing model %s configuration", modelType)
	configJson := map[string]any{
		"title":   title,
		"fields":  fields,
		"actions": actions,
		"apiUrl":  apiUrl,
	}

	modelConfig[strings.ToLower(modelType.Name())] = configJson
	typeRegistry[modelType.Name()] = func() interface{} { return model }
}

func getFieldConfigValue(fieldConfiguration string, configPrefix string) (string, bool) {
	fieldConfigParts := strings.Split(fieldConfiguration, ",")
	for i := range fieldConfigParts {
		if strings.Contains(fieldConfigParts[i], configPrefix) {
			if configValue, ok := strings.CutPrefix(fieldConfigParts[i], configPrefix); ok {
				return configValue, true
			}
		}
	}
	return "", false
}

func getModelConfig(modelType string) *map[string]interface{} {
	config := modelConfig[strings.ToLower(modelType)]
	return &config
}

func getModel(modelType string) (interface{}, error) {
	if factory, ok := typeRegistry[modelType]; ok {
		return factory(), nil
	}
	return nil, fmt.Errorf("type %s not found in registry", modelType)
}
