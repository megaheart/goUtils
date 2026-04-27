package appHelper

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// Setup name for application parameter by attribute tag `appParam:"<name>"` in struct field.
// Example:
//
//	type AppParamSetup struct {
//	    MaxUploadSize int    `appParam:"max-upload-size" appParamShorthand:"m" appParamUsage:"max upload size in bytes"`
//	    LogLevel      string `appParam:"log-level" appParamShorthand:"m" appParamUsage:"log level for the application"`
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
	param, _, err := ParseParamsAndSubCommands(defaultValue, envVarNamePrefix, nil)
	return param, err
}

type FlagAndEnvVarInfo struct {
	FlagName    string
	EnvVarName  string
	Description string
}

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
func ParseParamsAndSubCommands[T any](defaultValue T, envVarNamePrefix string, overwriteHelpFunc func(info []FlagAndEnvVarInfo)) (*T, []string, error) {
	commands := make([]string, 0)
	programName := os.Args[0]
	if idx := strings.LastIndex(programName, string(os.PathSeparator)); idx != -1 {
		programName = programName[idx+1:]
	}
	rootCmd := &cobra.Command{
		Use:   programName + " [command]",
		Short: "Demo cobra + pflag",
		Run: func(cmd *cobra.Command, _commands []string) {
			commands = _commands
		},
	}

	result := &defaultValue
	t := reflect.TypeOf(*result)
	// flagList := make([]interface{}, 0)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		paramName := f.Tag.Get("appParam")
		if paramName != "" {
			fv := reflect.ValueOf(result).Elem().FieldByName(f.Name)
			paramShorthand := f.Tag.Get("appParamShorthand")
			paramUsage := f.Tag.Get("appParamUsage")
			envVarName := getEnvVarNameFromAppParamName(paramName, envVarNamePrefix)
			if fv.IsValid() && fv.CanSet() {
				switch fv.Kind() {
				// String
				case reflect.String:
					defaultValue := fv.Interface().(string)
					if paramValue, ok := os.LookupEnv(envVarName); ok {
						fv.SetString(paramValue)
						rootCmd.Flags().StringP(paramName, paramShorthand, defaultValue, paramUsage)
						continue
					}
					pointer := fv.Addr().Interface().(*string)
					rootCmd.Flags().StringVarP(pointer, paramName, paramShorthand, defaultValue, paramUsage)
				// Int types
				case reflect.Int:
					defaultValue := fv.Interface().(int)
					if paramValue, ok := os.LookupEnv(envVarName); ok {
						intVal, err := strconv.ParseInt(paramValue, 10, 64)
						if err != nil {
							return nil, nil, fmt.Errorf("invalid int value for param '%s': '%s'", paramName, paramValue)
						}
						fv.SetInt(intVal)
						rootCmd.Flags().IntP(paramName, paramShorthand, defaultValue, paramUsage)
						continue
					}
					pointer := fv.Addr().Interface().(*int)
					rootCmd.Flags().IntVarP(pointer, paramName, paramShorthand, defaultValue, paramUsage)
				case reflect.Int8:
					defaultValue := fv.Interface().(int8)
					if paramValue, ok := os.LookupEnv(envVarName); ok {
						intVal, err := strconv.ParseInt(paramValue, 10, 8)
						if err != nil {
							return nil, nil, fmt.Errorf("invalid int8 value for param '%s': '%s'", paramName, paramValue)
						}
						fv.SetInt(intVal)
						rootCmd.Flags().Int8P(paramName, paramShorthand, defaultValue, paramUsage)
						continue
					}
					pointer := fv.Addr().Interface().(*int8)
					rootCmd.Flags().Int8VarP(pointer, paramName, paramShorthand, defaultValue, paramUsage)
				case reflect.Int16:
					defaultValue := fv.Interface().(int16)
					if paramValue, ok := os.LookupEnv(envVarName); ok {
						intVal, err := strconv.ParseInt(paramValue, 10, 16)
						if err != nil {
							return nil, nil, fmt.Errorf("invalid int16 value for param '%s': '%s'", paramName, paramValue)
						}
						fv.SetInt(intVal)
						rootCmd.Flags().Int16P(paramName, paramShorthand, defaultValue, paramUsage)
						continue
					}
					pointer := fv.Addr().Interface().(*int16)
					rootCmd.Flags().Int16VarP(pointer, paramName, paramShorthand, defaultValue, paramUsage)
				case reflect.Int32:
					defaultValue := fv.Interface().(int32)
					if paramValue, ok := os.LookupEnv(envVarName); ok {
						intVal, err := strconv.ParseInt(paramValue, 10, 32)
						if err != nil {
							return nil, nil, fmt.Errorf("invalid int32 value for param '%s': '%s'", paramName, paramValue)
						}
						fv.SetInt(intVal)
						rootCmd.Flags().Int32P(paramName, paramShorthand, defaultValue, paramUsage)
						continue
					}
					pointer := fv.Addr().Interface().(*int32)
					rootCmd.Flags().Int32VarP(pointer, paramName, paramShorthand, defaultValue, paramUsage)
				case reflect.Int64:
					defaultValue := fv.Interface().(int64)
					if paramValue, ok := os.LookupEnv(envVarName); ok {
						intVal, err := strconv.ParseInt(paramValue, 10, 64)
						if err != nil {
							return nil, nil, fmt.Errorf("invalid int64 value for param '%s': '%s'", paramName, paramValue)
						}
						fv.SetInt(intVal)
						rootCmd.Flags().Int64P(paramName, paramShorthand, defaultValue, paramUsage)
						continue
					}
					pointer := fv.Addr().Interface().(*int64)
					rootCmd.Flags().Int64VarP(pointer, paramName, paramShorthand, defaultValue, paramUsage)
				// Float types
				case reflect.Float32:
					defaultValue := fv.Interface().(float32)
					if paramValue, ok := os.LookupEnv(envVarName); ok {
						floatVal, err := strconv.ParseFloat(paramValue, 32)
						if err != nil {
							return nil, nil, fmt.Errorf("invalid float32 value for param '%s': '%s'", paramName, paramValue)
						}
						fv.SetFloat(floatVal)
						rootCmd.Flags().Float32P(paramName, paramShorthand, defaultValue, paramUsage)
						continue
					}
					pointer := fv.Addr().Interface().(*float32)
					rootCmd.Flags().Float32VarP(pointer, paramName, paramShorthand, defaultValue, paramUsage)

				case reflect.Float64:
					defaultValue := fv.Interface().(float64)
					if paramValue, ok := os.LookupEnv(envVarName); ok {
						floatVal, err := strconv.ParseFloat(paramValue, 64)
						if err != nil {
							return nil, nil, fmt.Errorf("invalid float64 value for param '%s': '%s'", paramName, paramValue)
						}
						fv.SetFloat(floatVal)
						rootCmd.Flags().Float64P(paramName, paramShorthand, defaultValue, paramUsage)
						continue
					}
					pointer := fv.Addr().Interface().(*float64)
					rootCmd.Flags().Float64VarP(pointer, paramName, paramShorthand, defaultValue, paramUsage)
				// Bool type
				case reflect.Bool:
					defaultValue := fv.Interface().(bool)
					if paramValue, ok := os.LookupEnv(envVarName); ok {
						boolVal, err := strconv.ParseBool(paramValue)
						if err != nil {
							return nil, nil, fmt.Errorf("invalid bool value for param '%s': '%s'", paramName, paramValue)
						}
						fv.SetBool(boolVal)
						rootCmd.Flags().BoolP(paramName, paramShorthand, defaultValue, paramUsage)
						continue
					}
					pointer := fv.Addr().Interface().(*bool)
					rootCmd.Flags().BoolVarP(pointer, paramName, paramShorthand, defaultValue, paramUsage)
				default:
					return nil, nil, fmt.Errorf("unsupported field type: %s", fv.Kind())
				}
			}
		}
	}

	rootCmd.Flags().SetInterspersed(true)

	var helpFunc func(cmd *cobra.Command, args []string)
	if overwriteHelpFunc != nil {
		helpFunc = func(cmd *cobra.Command, args []string) {
			infoList := make([]FlagAndEnvVarInfo, 0)
			for i := 0; i < t.NumField(); i++ {
				f := t.Field(i)
				paramName := f.Tag.Get("appParam")

				if paramName != "" {
					paramUsage := f.Tag.Get("appParamUsage")
					flagName := "--" + paramName
					shorthand := f.Tag.Get("appParamShorthand")
					if shorthand != "" {
						flagName = fmt.Sprintf("-%s, %s", shorthand, flagName)
					}
					envName := getEnvVarNameFromAppParamName(paramName, envVarNamePrefix)
					infoList = append(infoList, FlagAndEnvVarInfo{
						FlagName:    flagName,
						EnvVarName:  envName,
						Description: paramUsage,
					})
				}
			}
			overwriteHelpFunc(infoList)
			os.Exit(0)
		}
	} else {
		helpFunc = func(cmd *cobra.Command, args []string) {
			fmt.Printf("Usage of `%s`:\n", programName)
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
					shorthand := f.Tag.Get("appParamShorthand")
					if shorthand != "" {
						flagName = fmt.Sprintf("-%s, %s", shorthand, flagName)
					}
					envName := getEnvVarNameFromAppParamName(paramName, envVarNamePrefix)
					usageTable.AppendRow(table.Row{flagName, envName, paramUsage})
				}
			}
			usageTable.Render()
			os.Exit(0)
		}
	}
	rootCmd.SetHelpFunc(helpFunc)

	rootCmd.SetUsageFunc(func(cmd *cobra.Command) error {
		helpFunc(cmd, nil)
		return nil
	})

	if err := rootCmd.Execute(); err != nil {
		return nil, nil, err
	}
	return result, commands, nil
}

func getEnvVarNameFromAppParamName(kebab string, envVarNamePrefix string) string {
	result := strings.ToUpper(kebab)
	result = strings.ReplaceAll(result, "-", "_")
	result = strings.ReplaceAll(result, ".", "__")
	result = envVarNamePrefix + result
	return result
}
