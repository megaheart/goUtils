package appHelper

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
)

// Setup name for application parameter by attribute tag `appParam:"<name>"` in struct field.
// Example:
//
//	type AppParamSetup struct {
//	    MaxUploadSize int    `appParam:"max-upload-size"`
//	    LogLevel      string `appParam:"log-level"`
//	}
//
// Then use ParseParams to get parameters values from flag args or
// env vars and fill AppParamSetup struct.
//
// App param name must be unique, formatted as kebab-case (e.g., "max-upload-size")
// This setup make program auto-parse flag arg (params' names with "--" prefix,
//
//	e.g., "--max-upload-size") or env vars  (param's names in uppercase with underscores
//
// and "GO_SERVER_" prefix, e.g., "GO_SERVER_MAX_UPLOAD_SIZE") for each param.
// Flag arg has higher priority than env var.
func ParseParams[T any](defaultValue T, envVarNamePrefix string) (*T, error) {
	result := &defaultValue
	t := reflect.TypeOf(*result)
	flagList := make([]*string, 0)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		paramName := f.Tag.Get("appParam")
		if paramName != "" {
			flagList = append(flagList, flag.String(paramName, "", ""))
		}
	}
	flag.Parse()

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		paramName := f.Tag.Get("appParam")
		if paramName != "" {
			// get value from flag
			paramValue := *flagList[i]
			if paramValue == "" {
				// try to get from env var if flag not set
				envVarName := getEnvVarNameFromAppParamName(paramName, envVarNamePrefix)
				paramValue = os.Getenv(envVarName)
			}

			if paramValue == "" {
				continue
			}

			fv := reflect.ValueOf(result).Elem().FieldByName(f.Name)
			if fv.IsValid() && fv.CanSet() {
				switch fv.Kind() {
				case reflect.String:
					fv.SetString(paramValue)
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					// convert string to int
					var intVal int64
					count, err := fmt.Sscanf(paramValue, "%d", &intVal)
					if err != nil || count != 1 {
						return nil, fmt.Errorf("invalid int value for param '%s': %s", paramName, paramValue)
					}
					fv.SetInt(intVal)
				case reflect.Bool:
					// convert string to bool
					var boolVal bool
					count, err := fmt.Sscanf(paramValue, "%t", &boolVal)
					if err != nil || count != 1 {
						return nil, fmt.Errorf("invalid bool value for param '%s': %s", paramName, paramValue)
					}
					fv.SetBool(boolVal)
				// Add other types if needed
				default:
					return nil, fmt.Errorf("unsupported field type: %s", fv.Kind())
				}
			}
		}
	}
	return result, nil
}

func getEnvVarNameFromAppParamName(kebab string, envVarNamePrefix string) string {
	result := strings.ToUpper(kebab)
	result = strings.ReplaceAll(result, "-", "_")
	result = envVarNamePrefix + result
	return result
}
