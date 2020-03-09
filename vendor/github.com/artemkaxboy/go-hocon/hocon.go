package hocon

import (
	"fmt"
	"github.com/artemkaxboy/configuration"
	"github.com/artemkaxboy/configuration/hocon"
	"os"
	"reflect"
	"strings"
)

const pathKey = "path"
const nodeKey = "node"
const defaultKey = "default"

var (
	int64type = reflect.TypeOf(int64(0))

	tagKeys = map[string]interface{}{pathKey: nil, nodeKey: nil, defaultKey: nil}
)

type fieldWrapper struct {
	inner  *reflect.StructField
	single reflect.Type
}

func (ptr *fieldWrapper) getType() reflect.Type {
	if ptr.inner != nil {
		return ptr.inner.Type
	}
	return ptr.single
}

// getPath is a facade method to call getPath with inner (StructField) or return empty string
// if there is no inner element.
func (ptr *fieldWrapper) getPath(parentPath string) (string, error) {
	if ptr.inner == nil {
		return "", nil
	}
	return getPath(parentPath, ptr.inner)
}

// getPath returns HOCON path for current element.
//
// There are a few methods to set it for each element:
//
// 1. Set path value in struct tag, then it will be taken as is
//
// 2. Set node value in struct tag, then it will be added to the parent path with '.' delimiter
//
// 3. Do not set any tag, then the name of struct field (as is) will be added to the parent path with '.' delimiter
func getPath(parentPath string, field *reflect.StructField) (string, error) {
	tagMap, err := mapTag(field.Tag)
	if err != nil {
		return "", err
	}

	if path, exists := tagMap[pathKey]; exists {
		return path, nil
	}

	if len(parentPath) > 0 {
		parentPath = parentPath + "."
	}

	if node, exists := tagMap[nodeKey]; exists {
		return parentPath + node, nil
	}

	return parentPath + field.Name, nil
}

// LoadConfigFile loads HOCON files parameters to given structure.
func LoadConfigFile(filename string, receiver interface{}) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Cannot parse config: Panic")
			panic(r)
		}
	}()
	if err := checkFileAccessibility(filename); err != nil {
		return fmt.Errorf("cannot read configuration file: %w", err)
	}
	config := configuration.LoadConfig(filename)
	return loadConfig(config, receiver)
}

// LoadConfigText parses given text as HOCON and loads parameters to given structure.
func LoadConfigText(text string, receiver interface{}) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Cannot parse config: Panic")
			panic(r)
		}
	}()
	config := configuration.ParseString(text)
	return loadConfig(config, receiver)
}

// loadConfig - is an entrypoint to a recursive function which walk through receiver structure to
// find and load needed parameters.
func loadConfig(config *configuration.Config, receiver interface{}) error {
	wrapper := &fieldWrapper{
		single: reflect.ValueOf(receiver).Elem().Type(),
	}
	return loadStruct("", wrapper, reflect.ValueOf(receiver), config)
}

// loadStruct recursively walk through receiver struct nested elements to fill them with the
// config data.
func loadStruct(parentPath string, field *fieldWrapper, fieldValue reflect.Value, config *configuration.Config) error {
	currentPath, err2 := field.getPath(parentPath)
	if err2 != nil {
		return err2
	}

	for i := 0; i < field.getType().NumField(); i++ {
		innerField := field.getType().Field(i)
		if innerField.Type.Kind() == reflect.Struct {
			wrapper := &fieldWrapper{inner: &innerField}
			if err := loadStruct(currentPath, wrapper, fieldValue.Elem().FieldByName(innerField.Name).Addr(), config); err != nil {
				return err
			}
		} else {
			if err := loadValue(currentPath, &innerField, fieldValue.Elem().FieldByName(innerField.Name).Addr(), config); err != nil {
				return err
			}
		}
	}
	return nil
}

// loadValue loads value from config to fieldValue. It's a terminal method for recursive cycle of loadStruct.
func loadValue(parentPath string, field *reflect.StructField, fieldValue reflect.Value, config *configuration.Config) error {
	tagMap, err := mapTag(field.Tag)
	if err != nil {
		return err
	}

	// it's impossible to get error here while the only way to get it is give an element with incorrect tag and
	// map tag is doing before this statement.
	currentPath, _ := getPath(parentPath, field)

	typ := fieldValue.Elem().Type()

	hasDefault := false
	rawDefault := ""
	if rawDefault, hasDefault = tagMap[defaultKey]; hasDefault {
		if typ.Kind() == reflect.Slice {
			return fmt.Errorf("slices do not support default value: %s [%s]", field.Name, field.Tag)
		}
	} else {
		if !config.HasPath(currentPath) {
			return fmt.Errorf("no value either default value provided for %s [%s]", field.Name, field.Tag)
		}
	}

	switch typ.Kind() {
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		return fmt.Errorf("cannot use uint. Use int8/16/32/64 instead for %s [%s]", field.Name, field.Tag)

	case reflect.Int:
		return fmt.Errorf("cannot use int. Use int32 or int64 explicitly instead for %s [%s]", field.Name, field.Tag)

	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64, reflect.Bool:

		var defaultValue *reflect.Value
		if hasDefault {
			// we must check the correctness of default value even if value is provided
			var err1 error
			defaultValue, err1 = parseType(typ, rawDefault)
			if err1 != nil {
				return fmt.Errorf("wrong default value for %s [%s]: %w", field.Name, field.Tag, err1)
			}
		}

		hoconValue := config.GetValue(currentPath)

		value, err := parseHoconValue(typ, hoconValue)
		if err != nil {
			return fmt.Errorf("wrong value for %s [%s]: %w", field.Name, field.Tag, err)
		}

		if value != nil {
			fieldValue.Elem().Set(*value)
			return nil
		}
		fieldValue.Elem().Set(*defaultValue)
		return nil

	case reflect.String:
		typedValue := config.GetString(currentPath)
		fieldValue.Elem().SetString(typedValue)

	case reflect.Slice:
		list := config.GetStringList(currentPath)
		value := reflect.ValueOf(list)
		typedValue, err1 := parseList(typ, value)
		if err1 != nil {
			return err1
		}
		fieldValue.Elem().Set(typedValue)

	default:
		return fmt.Errorf("unimplemented data type %s", typ.Kind().String())
	}
	return nil
}

func parseType(typ reflect.Type, stringValue string) (*reflect.Value, error) {
	conf := configuration.ParseString(fmt.Sprintf("{k:%s}", stringValue))
	return parseHoconValue(typ, conf.GetValue("k"))
}

// parseHoconValue parses given hoconValue according to given reflect.Type and returns reflect.Value of this type.
func parseHoconValue(typ reflect.Type, hoconValue *hocon.HoconValue) (*reflect.Value, error) {
	if hoconValue == nil {
		return nil, nil
	}

	value, err := getExpandedValueSafely(typ, hoconValue)
	if err != nil {
		return nil, err
	}

	value, err = getTargetTypeValue(typ, value)
	if err != nil {
		return nil, err
	}

	return value, nil
}

// getTargetTypeValue scale down to float32, int32, int16 etc.
func getTargetTypeValue(typ reflect.Type, value *reflect.Value) (*reflect.Value, error) {
	switch typ.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32:
		targetTypeValue := value.Convert(typ)
		expected := value.Interface().(int64)
		actual := targetTypeValue.Convert(int64type).Interface().(int64)
		if expected != actual {
			return nil, fmt.Errorf("hocon: value out of range")
		}
		return &targetTypeValue, nil

	case reflect.Float32:
		targetTypeValue := value.Convert(typ)
		return &targetTypeValue, nil

	}

	return value, nil
}

// getExpandedValueSafely returns 64 bit value of ints and floats
// and regular value of others
func getExpandedValueSafely(typ reflect.Type, hoconValue *hocon.HoconValue) (*reflect.Value, error) {
	var value interface{}
	var err error

	switch typ.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value, err = hoconValue.GetInt64Safely()
		if err != nil {
			return nil, err
		}
	case reflect.Float32, reflect.Float64:
		value, err = hoconValue.GetFloat64Safely()
		if err != nil {
			return nil, err
		}
	case reflect.Bool:
		value, err = hoconValue.GetBooleanSafely()
		if err != nil {
			return nil, err
		}
	case reflect.String:
		value = hoconValue.GetString()
	}
	reflectValue := reflect.ValueOf(value)
	return &reflectValue, nil
}

// parseList parses given slice's values according to given reflect.Type and
// returns reflect.Value of slice of this type.
func parseList(typ reflect.Type, listValue reflect.Value) (reflect.Value, error) {
	sliceValue := reflect.MakeSlice(typ, listValue.Len(), listValue.Cap())
	for i := 0; i < listValue.Len(); i++ {
		res, err := parseType(typ.Elem(), listValue.Index(i).Interface().(string))
		if err != nil {
			return reflect.Value{}, err
		}
		sliceValue.Index(i).Set(*res)
	}
	return sliceValue, nil
}

// mapTag parses StructTag to aux Tag struct.
func mapTag(structTag reflect.StructTag) (map[string]string, error) {
	stringTag := structTag.Get("hocon")
	tagMap := make(map[string]string)
	if stringTag != "" {
		for _, item := range strings.Split(stringTag, ",") {
			pair := strings.Split(item, "=")
			if len(pair) != 2 {
				return nil, fmt.Errorf("tag format error: %s", stringTag)
			}
			key, value := pair[0], pair[1]

			if _, exists := tagKeys[key]; exists {
				tagMap[key] = value
			}
		}
	}
	return tagMap, nil
}

// checkFileAccessibility checks if a file accessible and is not a directory before we
// try using it to prevent further errors.
func checkFileAccessibility(filename string) error {
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}
	if info.Mode()&(1<<8) == 0 {
		return fmt.Errorf("%s permission denied", filename)
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory", filename)
	}
	return nil
}
