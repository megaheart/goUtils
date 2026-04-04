package appHelper

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

// Setup name for application parameter by attribute tag `appParam:"<name>"` in struct field.
// Example:
//
//	type AppParamSetup struct {
//	    MaxUploadSize int    `appParam:"max-upload-size" appParamUsage:"max upload size in bytes"`
//	    LogLevel      string `appParam:"log-level" appParamUsage:"log level for the application"`
//	}
//
// Then use ParseParams to get parameters values from flag args or
// env vars and fill AppParamSetup struct. Environment variable is more priority than flag arg.
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
	flagList := make([]interface{}, 0)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		paramName := f.Tag.Get("appParam")
		if paramName != "" {
			fv := reflect.ValueOf(result).Elem().FieldByName(f.Name)
			if fv.IsValid() && fv.CanSet() {
				switch fv.Kind() {
				// String
				case reflect.String:
					defaultValue := fv.Interface().(string)
					flagList = append(flagList, flag.String(paramName, defaultValue, f.Tag.Get("appParamUsage")))
				// Int types
				case reflect.Int:
					defaultValue := fv.Interface().(int)
					flagList = append(flagList, flag.Int64(paramName, int64(defaultValue), f.Tag.Get("appParamUsage")))
				case reflect.Int8:
					defaultValue := fv.Interface().(int8)
					flagList = append(flagList, flag.Int64(paramName, int64(defaultValue), f.Tag.Get("appParamUsage")))
				case reflect.Int16:
					defaultValue := fv.Interface().(int16)
					flagList = append(flagList, flag.Int64(paramName, int64(defaultValue), f.Tag.Get("appParamUsage")))
				case reflect.Int32:
					defaultValue := fv.Interface().(int32)
					flagList = append(flagList, flag.Int64(paramName, int64(defaultValue), f.Tag.Get("appParamUsage")))
				case reflect.Int64:
					defaultValue := fv.Interface().(int64)
					flagList = append(flagList, flag.Int64(paramName, defaultValue, f.Tag.Get("appParamUsage")))
				// Float types
				case reflect.Float32:
					defaultValue := fv.Interface().(float32)
					flagList = append(flagList, flag.Float64(paramName, float64(defaultValue), f.Tag.Get("appParamUsage")))
				case reflect.Float64:
					defaultValue := fv.Interface().(float64)
					flagList = append(flagList, flag.Float64(paramName, defaultValue, f.Tag.Get("appParamUsage")))
				// Bool type
				case reflect.Bool:
					defaultValue := fv.Interface().(bool)
					flagList = append(flagList, flag.Bool(paramName, defaultValue, f.Tag.Get("appParamUsage")))
				default:
					return nil, fmt.Errorf("unsupported field type: %s", fv.Kind())
				}
			}
		}
	}

	wantHelp := flag.Bool("help", false, "show help message")
	flag.Parse()

	if *wantHelp {
		fmt.Printf("Usage of `%s`:\n", os.Args[0])
		usageTable := table.NewWriter()
		usageTable.SetOutputMirror(os.Stdout)
		style := table.StyleDefault
		style.Options.DrawBorder = false
		style.Options.SeparateColumns = false
		style.Options.SeparateHeader = false
		style.Options.SeparateRows = false
		usageTable.SetStyle(style)

		usageTable.AppendRow(table.Row{"Flag", "Environment_Variable", "Description"})
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			paramName := f.Tag.Get("appParam")

			if paramName != "" {
				paramUsage := f.Tag.Get("appParamUsage")
				flagName := "--" + paramName
				envName := getEnvVarNameFromAppParamName(paramName, envVarNamePrefix)
				usageTable.AppendRow(table.Row{flagName, envName, paramUsage})
			}
		}
		usageTable.Render()
		os.Exit(0)
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		paramName := f.Tag.Get("appParam")
		if paramName != "" {
			envVarName := getEnvVarNameFromAppParamName(paramName, envVarNamePrefix)
			fv := reflect.ValueOf(result).Elem().FieldByName(f.Name)
			if fv.IsValid() && fv.CanSet() {
				switch fv.Kind() {
				case reflect.String:
					if paramValue, ok := os.LookupEnv(envVarName); ok {
						fv.SetString(paramValue)
						continue
					}
					paramValueRef := flagList[i].(*string)
					if paramValueRef != nil {
						fv.SetString(*paramValueRef)
					}
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					if paramValue, ok := os.LookupEnv(envVarName); ok {
						intVal, err := strconv.ParseInt(paramValue, 10, 64)
						if err != nil {
							return nil, fmt.Errorf("invalid int value for param '%s': '%s'", paramName, paramValue)
						}
						fv.SetInt(intVal)
						continue
					}
					paramValueRef := flagList[i].(*int64)
					if paramValueRef != nil {
						fv.SetInt(*paramValueRef)
					}
				case reflect.Float32, reflect.Float64:
					if paramValue, ok := os.LookupEnv(envVarName); ok {
						floatVal, err := strconv.ParseFloat(paramValue, 64)
						if err != nil {
							return nil, fmt.Errorf("invalid float value for param '%s': '%s'", paramName, paramValue)
						}
						fv.SetFloat(floatVal)
						continue
					}
					paramValueRef := flagList[i].(*float64)
					if paramValueRef != nil {
						fv.SetFloat(*paramValueRef)
					}
				case reflect.Bool:
					if paramValue, ok := os.LookupEnv(envVarName); ok {
						// convert string to bool
						boolVal, err := strconv.ParseBool(paramValue)
						if err != nil {
							return nil, fmt.Errorf("invalid bool value for param '%s': '%s'", paramName, paramValue)
						}
						fv.SetBool(boolVal)
						continue
					}
					paramValueRef := flagList[i].(*bool)
					if paramValueRef != nil {
						fv.SetBool(*paramValueRef)
					}
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
