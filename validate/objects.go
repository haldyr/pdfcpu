package validate

import (
	"fmt"

	"github.com/hhrutter/pdfcpu/types"
	"github.com/pkg/errors"
)

const (

	// REQUIRED is used for required dict entries.
	REQUIRED = true

	// OPTIONAL is used for optional dict entries.
	OPTIONAL = false
)

func validateEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName, entryName string, required bool, sinceVersion types.PDFVersion) (interface{}, error) {

	obj, found := dict.Find(entryName)
	if !found || obj == nil {
		if required {
			return nil, errors.Errorf("dict=%s required entry=%s missing.", dictName, entryName)
		}
		return nil, nil
	}

	obj, err := xRefTable.Dereference(obj)
	if err != nil {
		return nil, err
	}

	if obj == nil {
		if required {
			return nil, errors.Errorf("dict=%s required entry=%s missing.", dictName, entryName)
		}
		return nil, nil
	}

	// Version check
	err = xRefTable.ValidateVersion(fmt.Sprintf("dict=%s entry=%s", dictName, entryName), sinceVersion)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func validateArrayEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName string, entryName string, required bool, sinceVersion types.PDFVersion, validate func(types.PDFArray) bool) (*types.PDFArray, error) {

	logInfoValidate.Printf("validateArrayEntry begin: entry=%s\n", entryName)

	obj, err := dict.Entry(dictName, entryName, required)
	if err != nil || obj == nil {
		return nil, err
	}

	obj, err = xRefTable.Dereference(obj)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		if required {
			return nil, errors.Errorf("validateArrayEntry: dict=%s required entry=%s is nil", dictName, entryName)
		}
		logInfoValidate.Printf("validateArrayEntry end: optional entry %s is nil\n", entryName)
		return nil, nil
	}

	// Version check
	err = xRefTable.ValidateVersion(fmt.Sprintf("dict=%s entry=%s", dictName, entryName), sinceVersion)
	if err != nil {
		return nil, err
	}

	arr, ok := obj.(types.PDFArray)
	if !ok {
		return nil, errors.Errorf("validateArrayEntry: dict=%s entry=%s invalid type", dictName, entryName)
	}

	// Validation
	if validate != nil && !validate(arr) {
		return nil, errors.Errorf("validateArrayEntry: dict=%s entry=%s invalid dict entry", dictName, entryName)
	}

	logInfoValidate.Printf("validateArrayEntry end: entry=%s\n", entryName)

	return &arr, nil
}

func validateBooleanEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName string, entryName string, required bool, sinceVersion types.PDFVersion, validate func(bool) bool) (*types.PDFBoolean, error) {

	logInfoValidate.Printf("validateBooleanEntry begin: entry=%s\n", entryName)

	obj, err := dict.Entry(dictName, entryName, required)
	if err != nil || obj == nil {
		return nil, err
	}

	obj, err = xRefTable.Dereference(obj)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		if required {
			return nil, errors.Errorf("validateBooleanEntry: dict=%s required entry=%s missing", dictName, entryName)
		}
		logInfoValidate.Printf("validateBooleanEntry end: entry %s is nil\n", entryName)
		return nil, nil
	}

	// Version check
	err = xRefTable.ValidateVersion(fmt.Sprintf("dict=%s entry=%s", dictName, entryName), sinceVersion)
	if err != nil {
		return nil, err
	}

	b, ok := obj.(types.PDFBoolean)
	if !ok {
		return nil, errors.Errorf("validateBooleanEntry: dict=%s entry=%s invalid type", dictName, entryName)
	}

	// Validation
	if validate != nil && !validate(b.Value()) {
		return nil, errors.Errorf("validateBooleanEntry: dict=%s entry=%s invalid name dict entry", dictName, entryName)
	}

	logInfoValidate.Printf("validateBooleanEntry end: entry=%s\n", entryName)

	return &b, nil
}

func validateBooleanArrayEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName string, entryName string, required bool, sinceVersion types.PDFVersion, validate func(types.PDFArray) bool) (*types.PDFArray, error) {

	logInfoValidate.Printf("validateBooleanArrayEntry begin: entry=%s\n", entryName)

	arr, err := validateArrayEntry(xRefTable, dict, dictName, entryName, required, sinceVersion, validate)
	if err != nil || arr == nil {
		return nil, err
	}

	for i, obj := range *arr {

		obj, err := xRefTable.Dereference(obj)
		if err != nil {
			return nil, err
		}
		if obj == nil {
			continue
		}

		_, ok := obj.(types.PDFBoolean)
		if !ok {
			return nil, errors.Errorf("validateBooleanArrayEntry: dict=%s entry=%s invalid type at index %d\n", dictName, entryName, i)
		}

	}

	logInfoValidate.Printf("validateBooleanArrayEntry end: entry=%s\n", entryName)

	return arr, nil
}

func validateDateObject(xRefTable *types.XRefTable, obj interface{}, sinceVersion types.PDFVersion) (types.PDFStringLiteral, error) {
	return xRefTable.DereferenceStringLiteral(obj, sinceVersion, validateDate)
}

func validateDateEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName string, entryName string, required bool, sinceVersion types.PDFVersion) (*types.PDFStringLiteral, error) {

	logInfoValidate.Printf("validateDateEntry begin: entry=%s\n", entryName)

	obj, err := dict.Entry(dictName, entryName, required)
	if err != nil || obj == nil {
		return nil, err
	}

	obj, err = xRefTable.Dereference(obj)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		if required {
			return nil, errors.Errorf("validateDateEntry: dict=%s required entry=%s is nil", dictName, entryName)
		}
		logInfoValidate.Printf("validateDateEntry end: optional entry %s is nil\n", entryName)
		return nil, nil
	}

	// Version check
	err = xRefTable.ValidateVersion(fmt.Sprintf("dict=%s entry=%s", dictName, entryName), sinceVersion)
	if err != nil {
		return nil, err
	}

	date, ok := obj.(types.PDFStringLiteral)
	if !ok {
		return nil, errors.Errorf("validateDateEntry: dict=%s entry=%s invalid type", dictName, entryName)
	}

	// Validation
	if ok := validateDate(date.Value()); !ok {
		return nil, errors.Errorf("validateDateEntry: dict=%s entry=%s invalid dict entry", dictName, entryName)
	}

	logInfoValidate.Printf("validateDateEntry end: entry=%s\n", entryName)

	return &date, nil
}

func validateDictEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName string, entryName string, required bool, sinceVersion types.PDFVersion, validate func(types.PDFDict) bool) (*types.PDFDict, error) {

	logInfoValidate.Printf("validateDictEntry begin: entry=%s\n", entryName)

	obj, err := dict.Entry(dictName, entryName, required)
	if err != nil || obj == nil {
		return nil, err
	}

	obj, err = xRefTable.Dereference(obj)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		if required {
			return nil, errors.Errorf("validateDictEntry: dict=%s required entry=%s is nil", dictName, entryName)
		}
		logInfoValidate.Printf("validateDictEntry end: optional entry %s is nil\n", entryName)
		return nil, nil
	}

	// Version check
	err = xRefTable.ValidateVersion(fmt.Sprintf("dict=%s entry=%s", dictName, entryName), sinceVersion)
	if err != nil {
		return nil, err
	}

	d, ok := obj.(types.PDFDict)
	if !ok {
		return nil, errors.Errorf("validateDictEntry: dict=%s entry=%s invalid type", dictName, entryName)
	}

	// Validation
	if validate != nil && !validate(d) {
		return nil, errors.Errorf("validateDictEntry: dict=%s entry=%s invalid dict entry", dictName, entryName)
	}

	logInfoValidate.Printf("validateDictEntry end: entry=%s\n", entryName)

	return &d, nil
}

func validateFloat(xRefTable *types.XRefTable, obj interface{}, validate func(float64) bool) (*types.PDFFloat, error) {

	logInfoValidate.Println("validateFloat begin")

	obj, err := xRefTable.Dereference(obj)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, errors.New("validateFloat: missing object")
	}

	f, ok := obj.(types.PDFFloat)
	if !ok {
		return nil, errors.New("validateFloat: invalid type")
	}

	// Validation
	if validate != nil && !validate(f.Value()) {
		return nil, errors.Errorf("validateFloat: invalid float: %s\n", f)
	}

	logInfoValidate.Println("validateFloat end")

	return &f, nil
}

func validateFloatEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName string, entryName string, required bool, sinceVersion types.PDFVersion, validate func(float64) bool) (*types.PDFFloat, error) {

	logInfoValidate.Printf("validateFloatEntry begin: entry=%s\n", entryName)

	obj, err := dict.Entry(dictName, entryName, required)
	if err != nil || obj == nil {
		return nil, err
	}

	obj, err = xRefTable.Dereference(obj)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		if required {
			return nil, errors.Errorf("validateFloatEntry: dict=%s required entry=%s is nil", dictName, entryName)
		}
		logInfoValidate.Printf("validateFloatEntry end: optional entry %s is nil\n", entryName)
		return nil, nil
	}

	// Version check
	err = xRefTable.ValidateVersion(fmt.Sprintf("dict=%s entry=%s", dictName, entryName), sinceVersion)
	if err != nil {
		return nil, err
	}

	f, ok := obj.(types.PDFFloat)
	if !ok {
		return nil, errors.Errorf("validateFloatEntry: dict=%s entry=%s invalid type", dictName, entryName)
	}

	// Validation
	if validate != nil && !validate(f.Value()) {
		return nil, errors.Errorf("validateFloatEntry: dict=%s entry=%s invalid dict entry", dictName, entryName)
	}

	logInfoValidate.Printf("validateFloatEntry end: entry=%s\n", entryName)

	return &f, nil
}

func validateFunctionEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName string, entryName string, required bool, sinceVersion types.PDFVersion) error {

	logInfoValidate.Printf("validateFunctionEntry begin: entry=%s\n", entryName)

	obj, err := dict.Entry(dictName, entryName, required)
	if err != nil || obj == nil {
		return err
	}

	err = xRefTable.ValidateVersion(fmt.Sprintf("dict=%s entry=%s", dictName, entryName), sinceVersion)
	if err != nil {
		return err
	}

	err = validateFunction(xRefTable, obj)
	if err != nil {
		return err
	}

	logInfoValidate.Printf("validateFunctionEntry end: entry=%s\n", entryName)

	return nil
}

func validateFunctionArrayEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName string, entryName string, required bool, sinceVersion types.PDFVersion, validate func(types.PDFArray) bool) (*types.PDFArray, error) {

	logInfoValidate.Printf("validateFunctionArrayEntry begin: entry=%s\n", entryName)

	arr, err := validateArrayEntry(xRefTable, dict, dictName, entryName, required, sinceVersion, validate)
	if err != nil || arr == nil {
		return nil, err
	}

	for _, obj := range *arr {
		err = validateFunction(xRefTable, obj)
		if err != nil {
			return nil, err
		}
	}

	logInfoValidate.Printf("validateFunctionArrayEntry end: entry=%s\n", entryName)

	return arr, nil
}

func validateFunctionOrArrayOfFunctionsEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName, entryName string, required bool, sinceVersion types.PDFVersion) error {

	logInfoValidate.Printf("validateFunctionOrArrayOfFunctionsEntry begin: entry=%s\n", entryName)

	obj, err := dict.Entry(dictName, entryName, required)
	if err != nil || obj == nil {
		return err
	}

	obj, err = xRefTable.Dereference(obj)
	if err != nil {
		return err
	}
	if obj == nil {
		if required {
			return errors.Errorf("validateFunctionOrArrayOfFunctionsEntry: dict=%s required entry=%s is nil", dictName, entryName)
		}
		logInfoValidate.Printf("validateFunctionOrArrayOfFunctionsEntry end: optional entry %s is nil\n", entryName)
		return nil
	}

	switch obj := obj.(type) {

	case types.PDFArray:

		for _, obj := range obj {

			if obj == nil {
				continue
			}

			err = validateFunction(xRefTable, obj)
			if err != nil {
				return err
			}

		}

	default:
		err = validateFunction(xRefTable, obj)
		if err != nil {
			return err
		}

	}

	err = xRefTable.ValidateVersion(fmt.Sprintf("dict=%s entry=%s", dictName, entryName), sinceVersion)
	if err != nil {
		return err
	}

	logInfoValidate.Printf("validateFunctionOrArrayOfFunctionsEntry end: entry=%s\n", entryName)

	return nil
}

func validateIndRefEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName string, entryName string, required bool, sinceVersion types.PDFVersion) (*types.PDFIndirectRef, error) {

	logInfoValidate.Printf("validateIndRefEntry begin: entry=%s\n", entryName)

	obj, err := dict.Entry(dictName, entryName, required)
	if err != nil || obj == nil {
		return nil, err
	}

	indRef, ok := obj.(types.PDFIndirectRef)
	if !ok {
		return nil, errors.Errorf("validateIndRefEntry: dict=%s entry=%s invalid type", dictName, entryName)
	}

	// Version check
	err = xRefTable.ValidateVersion(fmt.Sprintf("dict=%s entry=%s", dictName, entryName), sinceVersion)
	if err != nil {
		return nil, err
	}

	logInfoValidate.Printf("validateIndRefEntry end: entry=%s\n", entryName)

	return &indRef, nil
}

func validateIndRefArrayEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName string, entryName string, required bool, sinceVersion types.PDFVersion, validate func(types.PDFArray) bool) (*types.PDFArray, error) {

	logInfoValidate.Printf("validateIndRefArrayEntry begin: entry=%s\n", entryName)

	arr, err := validateArrayEntry(xRefTable, dict, dictName, entryName, required, sinceVersion, validate)
	if err != nil || arr == nil {
		return nil, err
	}

	for i, obj := range *arr {
		_, ok := obj.(types.PDFIndirectRef)
		if !ok {
			return nil, errors.Errorf("validateIndRefArrayEntry: invalid type at index %d\n", i)
		}
	}

	logInfoValidate.Printf("validateIndRefArrayEntry end: entry=%s \n", entryName)

	return arr, nil
}

func validateInteger(xRefTable *types.XRefTable, obj interface{}, validate func(int) bool) (*types.PDFInteger, error) {

	logInfoValidate.Println("validateInteger begin")

	obj, err := xRefTable.Dereference(obj)
	if err != nil {
		return nil, err
	}

	if obj == nil {
		return nil, errors.New("validateInteger: missing object")
	}

	i, ok := obj.(types.PDFInteger)
	if !ok {
		return nil, errors.New("validateInteger: invalid type")
	}

	// Validation
	if validate != nil && !validate(i.Value()) {
		return nil, errors.Errorf("validateInteger: invalid integer: %s\n", i)
	}

	logInfoValidate.Println("validateInteger end")

	return &i, nil
}

func validateIntegerEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName string, entryName string, required bool, sinceVersion types.PDFVersion, validate func(int) bool) (*types.PDFInteger, error) {

	logInfoValidate.Printf("validateIntegerEntry begin: entry=%s\n", entryName)

	obj, err := dict.Entry(dictName, entryName, required)
	if err != nil || obj == nil {
		return nil, err
	}

	obj, err = xRefTable.Dereference(obj)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		if required {
			return nil, errors.Errorf("validateIntegerEntry: dict=%s required entry=%s is nil", dictName, entryName)
		}
		logInfoValidate.Printf("validateIntegerEntry end: optional entry %s is nil\n", entryName)
		return nil, nil
	}

	// Version check
	err = xRefTable.ValidateVersion(fmt.Sprintf("dict=%s entry=%s", dictName, entryName), sinceVersion)
	if err != nil {
		return nil, err
	}

	i, ok := obj.(types.PDFInteger)
	if !ok {
		return nil, errors.Errorf("validateIntegerEntry: dict=%s entry=%s invalid type", dictName, entryName)
	}

	// Validation
	if validate != nil && !validate(i.Value()) {
		return nil, errors.Errorf("validateIntegerEntry: dict=%s entry=%s invalid dict entry", dictName, entryName)
	}

	logInfoValidate.Printf("validateIntegerEntry end: entry=%s\n", entryName)

	return &i, nil
}

func validateIntegerArray(xRefTable *types.XRefTable, obj interface{}) (*types.PDFArray, error) {

	logInfoValidate.Println("validateIntegerArray begin")

	a, err := xRefTable.DereferenceArray(obj)
	if err != nil || a == nil {
		return nil, err
	}

	for i, obj := range *a {

		obj, err = xRefTable.Dereference(obj)
		if err != nil {
			return nil, err
		}

		if obj == nil {
			continue
		}

		switch obj.(type) {

		case types.PDFInteger:
			// no further processing.

		default:
			return nil, errors.Errorf("validateIntegerArray: invalid type at index %d\n", i)
		}

	}

	logInfoValidate.Println("validateIntegerArray end")

	return a, nil
}

func validateIntegerArrayEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName string, entryName string, required bool, sinceVersion types.PDFVersion, validate func(types.PDFArray) bool) (*types.PDFArray, error) {

	logInfoValidate.Printf("validateIntegerArrayEntry begin: entry=%s\n", entryName)

	arr, err := validateArrayEntry(xRefTable, dict, dictName, entryName, required, sinceVersion, validate)
	if err != nil || arr == nil {
		return nil, err
	}

	for i, obj := range *arr {

		obj, err = xRefTable.Dereference(obj)
		if err != nil {
			return nil, err
		}

		if obj == nil {
			continue
		}

		_, ok := obj.(types.PDFInteger)
		if !ok {
			return nil, errors.Errorf("validateIntegerArrayEntry: dict=%s entry=%s invalid type at index %d\n", dictName, entryName, i)
		}

	}

	logInfoValidate.Printf("validateIntegerArrayEntry end: entry=%s\n", entryName)

	return arr, nil
}

func validateName(xRefTable *types.XRefTable, obj interface{}, validate func(string) bool) (*types.PDFName, error) {

	logInfoValidate.Println("validateName begin")

	obj, err := xRefTable.Dereference(obj)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, errors.New("validateName: missing object")
	}

	name, ok := obj.(types.PDFName)
	if !ok {
		return nil, errors.New("validateName: invalid type")
	}

	// Validation
	if validate != nil && !validate(name.String()) {
		return nil, errors.Errorf("validateName: invalid name: %s\n", name)
	}

	logInfoValidate.Println("validateName end")

	return &name, nil
}

func validateNameEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName string, entryName string, required bool, sinceVersion types.PDFVersion, validate func(string) bool) (*types.PDFName, error) {

	logInfoValidate.Printf("validateNameEntry begin: entry=%s\n", entryName)

	obj, err := dict.Entry(dictName, entryName, required)
	if err != nil || obj == nil {
		return nil, err
	}

	obj, err = xRefTable.Dereference(obj)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		if required {
			return nil, errors.Errorf("validateNameEntry: dict=%s required entry=%s is nil", dictName, entryName)
		}
		logInfoValidate.Printf("validateNameEntry end: optional entry %s is nil\n", entryName)
		return nil, nil
	}

	// Version check
	err = xRefTable.ValidateVersion(fmt.Sprintf("dict=%s entry=%s", dictName, entryName), sinceVersion)
	if err != nil {
		return nil, err
	}

	name, ok := obj.(types.PDFName)
	if !ok {
		return nil, errors.Errorf("validateNameEntry: dict=%s entry=%s invalid type", dictName, entryName)
	}

	// Validation
	if validate != nil && !validate(name.String()) {
		return nil, errors.Errorf("validateNameEntry: dict=%s entry=%s invalid dict entry: %s", dictName, entryName, name.String())
	}

	logInfoValidate.Printf("validateNameEntry end: entry=%s\n", entryName)

	return &name, nil
}

func validateNameArray(xRefTable *types.XRefTable, obj interface{}) (*types.PDFArray, error) {

	logInfoValidate.Println("validateNameArray begin")

	arr, err := xRefTable.DereferenceArray(obj)
	if err != nil || arr == nil {
		return nil, err
	}

	for i, obj := range *arr {

		obj, err = xRefTable.Dereference(obj)
		if err != nil {
			return nil, err
		}

		if obj == nil {
			continue
		}

		_, ok := obj.(types.PDFName)
		if !ok {
			return nil, errors.Errorf("validateNameArray: invalid type at index %d\n", i)
		}

	}

	logInfoValidate.Println("validateNameArray end")

	return arr, nil
}

func validateNameArrayEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName string, entryName string, required bool, sinceVersion types.PDFVersion, validate func(a types.PDFArray) bool) (*types.PDFArray, error) {

	logInfoValidate.Printf("validateNameArrayEntry begin: entry=%s\n", entryName)

	arr, err := validateArrayEntry(xRefTable, dict, dictName, entryName, required, sinceVersion, validate)
	if err != nil || arr == nil {
		return nil, err
	}

	for i, obj := range *arr {

		obj, err = xRefTable.Dereference(obj)
		if err != nil {
			return nil, err
		}

		if obj == nil {
			continue
		}

		_, ok := obj.(types.PDFName)
		if !ok {
			return nil, errors.Errorf("validateNameArrayEntry: dict=%s entry=%s invalid type at index %d\n", dictName, entryName, i)
		}

	}

	logInfoValidate.Printf("validateNameArrayEntry end: entry=%s\n", entryName)

	return arr, nil
}

func validateNumber(xRefTable *types.XRefTable, obj interface{}) (interface{}, error) {

	logInfoValidate.Println("validateNumber begin")

	n, err := xRefTable.Dereference(obj)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, errors.New("validateNumber: missing object")
	}

	switch n.(type) {

	case types.PDFInteger:
		// no further processing.

	case types.PDFFloat:
		// no further processing.

	default:
		return nil, errors.New("validateNumber: invalid type")

	}

	logInfoValidate.Println("validateNumber end ")

	return n, nil
}

func validateNumberEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName string, entryName string, required bool, sinceVersion types.PDFVersion, validate func(f float64) bool) (interface{}, error) {

	logInfoValidate.Printf("validateNumberEntry begin: entry=%s\n", entryName)

	obj, err := dict.Entry(dictName, entryName, required)
	if err != nil || obj == nil {
		return nil, err
	}

	// Version check
	err = xRefTable.ValidateVersion(fmt.Sprintf("dict=%s entry=%s", dictName, entryName), sinceVersion)
	if err != nil {
		return nil, err
	}

	obj, err = validateNumber(xRefTable, obj)
	if err != nil {
		return nil, err
	}

	var f float64

	// Validation
	switch o := obj.(type) {

	case types.PDFInteger:
		f = float64(o.Value())

	case types.PDFFloat:
		f = o.Value()
	}

	if validate != nil && !validate(f) {
		return nil, errors.Errorf("validateFloatEntry: dict=%s entry=%s invalid dict entry", dictName, entryName)
	}

	logInfoValidate.Printf("validateNumberEntry end: entry=%s\n", entryName)

	return obj, nil
}

func validateNumberArray(xRefTable *types.XRefTable, obj interface{}) (*types.PDFArray, error) {

	logInfoValidate.Println("validateNumberArray begin")

	arrp, err := xRefTable.DereferenceArray(obj)
	if err != nil || arrp == nil {
		return nil, err
	}

	for i, obj := range *arrp {

		obj, err = xRefTable.Dereference(obj)
		if err != nil {
			return nil, err
		}

		if obj == nil {
			continue
		}

		switch obj.(type) {

		case types.PDFInteger:
			// no further processing.

		case types.PDFFloat:
			// no further processing.

		default:
			return nil, errors.Errorf("validateNumberArray: invalid type at index %d\n", i)
		}

	}

	logInfoValidate.Println("validateNumberArray end")

	return nil, err
}

func validateNumberArrayEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName string, entryName string, required bool, sinceVersion types.PDFVersion, validate func(types.PDFArray) bool) (*types.PDFArray, error) {

	logInfoValidate.Printf("validateNumberArrayEntry begin: entry=%s\n", entryName)

	arr, err := validateArrayEntry(xRefTable, dict, dictName, entryName, required, sinceVersion, validate)
	if err != nil || arr == nil {
		return nil, err
	}

	for i, obj := range *arr {

		obj, err = xRefTable.Dereference(obj)
		if err != nil {
			return nil, err
		}

		if obj == nil {
			continue
		}

		switch obj.(type) {

		case types.PDFInteger:
			// no further processing.

		case types.PDFFloat:
			// no further processing.

		default:
			return nil, errors.Errorf("validateNumberArrayEntry: invalid type at index %d\n", i)
		}

	}

	logInfoValidate.Printf("validateNumberArrayEntry end: entry=%s\n", entryName)

	return arr, nil
}

func validateRectangleEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName string, entryName string, required bool, sinceVersion types.PDFVersion, validate func(types.PDFArray) bool) (*types.PDFArray, error) {

	logInfoValidate.Printf("validateRectangleEntry begin: entry=%s\n", entryName)

	arr, err := validateNumberArrayEntry(xRefTable, dict, dictName, entryName, required, sinceVersion, func(arr types.PDFArray) bool { return len(arr) == 4 })
	if err != nil || arr == nil {
		return nil, err
	}

	if validate != nil && !validate(*arr) {
		return nil, errors.Errorf("validateRectangleEntry: dict=%s entry=%s invalid rectangle entry", dictName, entryName)
	}

	logInfoValidate.Printf("validateRectangleEntry end: entry=%s\n", entryName)

	return arr, nil
}

func validateStreamDict(xRefTable *types.XRefTable, obj interface{}) (*types.PDFStreamDict, error) {

	logInfoValidate.Println("validateStreamDict begin")

	obj, err := xRefTable.Dereference(obj)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, errors.New("validateStreamDict: missing object")
	}

	sd, ok := obj.(types.PDFStreamDict)
	if !ok {
		return nil, errors.New("validateStreamDict: invalid type")
	}

	logInfoValidate.Println("validateStreamDict endobj")

	return &sd, nil
}

func validateStreamDictEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName string, entryName string, required bool, sinceVersion types.PDFVersion, validate func(types.PDFStreamDict) bool) (*types.PDFStreamDict, error) {

	logInfoValidate.Printf("validateStreamDictEntry begin: entry=%s\n", entryName)

	obj, err := dict.Entry(dictName, entryName, required)
	if err != nil || obj == nil {
		return nil, err
	}

	obj, err = xRefTable.Dereference(obj)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		if required {
			return nil, errors.Errorf("validateStreamDictEntry: dict=%s required entry=%s is nil", dictName, entryName)
		}
		logInfoValidate.Printf("validateStreamDictEntry end: optional entry %s is nil\n", entryName)
		return nil, nil
	}

	// Version check
	err = xRefTable.ValidateVersion(fmt.Sprintf("dict=%s entry=%s", dictName, entryName), sinceVersion)
	if err != nil {
		return nil, err
	}

	sd, ok := obj.(types.PDFStreamDict)
	if !ok {
		return nil, errors.Errorf("validateStreamDictEntry: dict=%s entry=%s invalid type", dictName, entryName)
	}

	// Validation
	if validate != nil && !validate(sd) {
		return nil, errors.Errorf("validateStreamDictEntry: dict=%s entry=%s invalid dict entry", dictName, entryName)
	}

	logInfoValidate.Printf("validateStreamDictEntry end: entry=%s\n", entryName)

	return &sd, nil
}

func validateString(xRefTable *types.XRefTable, obj interface{}, validate func(string) bool) (*string, error) {

	logInfoValidate.Println("validateString begin")

	obj, err := xRefTable.Dereference(obj)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, errors.New("validateString: missing object")
	}

	var s string

	switch obj := obj.(type) {

	case types.PDFStringLiteral:
		s = obj.Value()

	case types.PDFHexLiteral:
		s = obj.Value()

	default:
		return nil, errors.New("validateString: invalid type")
	}

	// Validation
	if validate != nil && !validate(s) {
		return nil, errors.Errorf("validateString: %s invalid", s)
	}

	logInfoValidate.Println("validateString end")

	return &s, nil
}

func validateStringEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName string, entryName string, required bool, sinceVersion types.PDFVersion, validate func(string) bool) (*string, error) {

	logInfoValidate.Printf("validateStringEntry begin: entry=%s\n", entryName)

	obj, err := dict.Entry(dictName, entryName, required)
	if err != nil || obj == nil {
		return nil, err
	}

	obj, err = xRefTable.Dereference(obj)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		if required {
			return nil, errors.Errorf("validateStringEntry: dict=%s required entry=%s is nil", dictName, entryName)
		}
		logInfoValidate.Printf("validateStringEntry end: optional entry %s is nil\n", entryName)
		return nil, nil
	}

	// Version check
	err = xRefTable.ValidateVersion(fmt.Sprintf("dict=%s entry=%s", dictName, entryName), sinceVersion)
	if err != nil {
		return nil, err
	}

	var s string

	switch obj := obj.(type) {

	case types.PDFStringLiteral:
		s = obj.Value()

	case types.PDFHexLiteral:
		s = obj.Value()

	default:
		return nil, errors.Errorf("validateStringEntry: dict=%s entry=%s invalid type", dictName, entryName)
	}

	// Validation
	if validate != nil && !validate(s) {
		return nil, errors.Errorf("validateStringEntry: dict=%s entry=%s invalid dict entry", dictName, entryName)
	}

	logInfoValidate.Printf("validateStringEntry end: entry=%s\n", entryName)

	return &s, nil
}

func validateStringArrayEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName string, entryName string, required bool, sinceVersion types.PDFVersion, validate func(types.PDFArray) bool) (*types.PDFArray, error) {

	logInfoValidate.Printf("validateStringArrayEntry begin: entry=%s\n", entryName)

	arr, err := validateArrayEntry(xRefTable, dict, dictName, entryName, required, sinceVersion, validate)
	if err != nil || arr == nil {
		return nil, err
	}

	for i, obj := range *arr {

		obj, err = xRefTable.Dereference(obj)
		if err != nil {
			return nil, err
		}

		if obj == nil {
			continue
		}

		switch obj.(type) {

		case types.PDFStringLiteral:
			// no further processing.

		case types.PDFHexLiteral:
			// no further processing

		default:
			return nil, errors.Errorf("validateStringArrayEntry: invalid type at index %d\n", i)
		}

	}

	logInfoValidate.Printf("validateStringArrayEntry end: entry=%s\n", entryName)

	return arr, nil
}

func validateArrayArrayEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName string, entryName string, required bool, sinceVersion types.PDFVersion, validate func(types.PDFArray) bool) (*types.PDFArray, error) {

	logInfoValidate.Printf("validateArrayArrayEntry begin: entry=%s\n", entryName)

	arr, err := validateArrayEntry(xRefTable, dict, dictName, entryName, required, sinceVersion, validate)
	if err != nil || arr == nil {
		return nil, err
	}

	for i, obj := range *arr {

		obj, err = xRefTable.Dereference(obj)
		if err != nil {
			return nil, err
		}

		if obj == nil {
			continue
		}

		switch obj.(type) {

		case types.PDFArray:
			// no further processing.

		default:
			return nil, errors.Errorf("validateArrayArrayEntry: invalid type at index %d\n", i)
		}

	}

	logInfoValidate.Printf("validateArrayArrayEntry end: entry=%s\n", entryName)

	return arr, nil
}

func validateStringOrStreamEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName, entryName string, required bool, sinceVersion types.PDFVersion) error {

	logInfoValidate.Printf("validateStringOrStreamEntry begin: entry=%s\n", entryName)

	obj, err := dict.Entry(dictName, entryName, required)
	if err != nil || obj == nil {
		return err
	}

	obj, err = xRefTable.Dereference(obj)
	if err != nil {
		return err
	}
	if obj == nil {
		if required {
			return errors.Errorf("validateStringOrStreamEntry: dict=%s required entry=%s is nil", dictName, entryName)
		}
		logInfoValidate.Printf("validateStringOrStreamEntry end: optional entry %s is nil\n", entryName)
		return nil
	}

	// Version check
	err = xRefTable.ValidateVersion(fmt.Sprintf("dict=%s entry=%s", dictName, entryName), sinceVersion)
	if err != nil {
		return err
	}

	switch obj.(type) {

	case types.PDFStringLiteral, types.PDFHexLiteral, types.PDFStreamDict:
		// no further processing

	default:
		return errors.Errorf("validateStringOrStreamEntry: dict=%s entry=%s invalid type", dictName, entryName)
	}

	logInfoValidate.Printf("validateStringOrStreamEntry end: entry=%s\n", entryName)

	return nil
}

func validateNameOrStringEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName, entryName string, required bool, sinceVersion types.PDFVersion) error {

	logInfoValidate.Printf("validateNameOrStringEntry begin: entry=%s\n", entryName)

	obj, err := dict.Entry(dictName, entryName, required)
	if err != nil || obj == nil {
		return err
	}

	obj, err = xRefTable.Dereference(obj)
	if err != nil {
		return err
	}
	if obj == nil {
		if required {
			return errors.Errorf("validateNameOrStringEntry: dict=%s required entry=%s is nil", dictName, entryName)
		}
		logInfoValidate.Printf("validateNameOrStringEntry end: optional entry %s is nil\n", entryName)
		return nil
	}

	// Version check
	err = xRefTable.ValidateVersion(fmt.Sprintf("dict=%s entry=%s", dictName, entryName), sinceVersion)
	if err != nil {
		return err
	}

	switch obj.(type) {

	case types.PDFStringLiteral, types.PDFName:
		// no further processing

	default:
		return errors.Errorf("validateNameOrStringEntry: dict=%s entry=%s invalid type", dictName, entryName)
	}

	logInfoValidate.Printf("validateNameOrStringEntry end: entry=%s\n", entryName)

	return nil
}

func validateIntOrStringEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName, entryName string, required bool, sinceVersion types.PDFVersion) error {

	logInfoValidate.Printf("validateIntOrStringEntry begin: entry=%s\n", entryName)

	obj, err := dict.Entry(dictName, entryName, required)
	if err != nil || obj == nil {
		return err
	}

	obj, err = xRefTable.Dereference(obj)
	if err != nil {
		return err
	}
	if obj == nil {
		if required {
			return errors.Errorf("validateIntOrStringEntry: dict=%s required entry=%s is nil", dictName, entryName)
		}
		logInfoValidate.Printf("validateIntOrStringEntry end: optional entry %s is nil\n", entryName)
		return nil
	}

	// Version check
	err = xRefTable.ValidateVersion(fmt.Sprintf("dict=%s entry=%s", dictName, entryName), sinceVersion)
	if err != nil {
		return err
	}

	switch obj.(type) {

	case types.PDFStringLiteral, types.PDFHexLiteral, types.PDFInteger:
		// no further processing

	default:
		return errors.Errorf("validateIntOrStringEntry: dict=%s entry=%s invalid type", dictName, entryName)
	}

	logInfoValidate.Printf("validateIntOrStringEntry end: entry=%s\n", entryName)

	return nil
}

func validateIntOrDictEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName, entryName string, required bool, sinceVersion types.PDFVersion) error {

	logInfoValidate.Printf("validateIntOrDictEntry begin: entry=%s\n", entryName)

	obj, err := dict.Entry(dictName, entryName, required)
	if err != nil || obj == nil {
		return err
	}

	obj, err = xRefTable.Dereference(obj)
	if err != nil {
		return err
	}
	if obj == nil {
		if required {
			return errors.Errorf("validateIntOrDictEntry: dict=%s required entry=%s is nil", dictName, entryName)
		}
		logInfoValidate.Printf("validateIntOrDictEntry end: optional entry %s is nil\n", entryName)
		return nil
	}

	// Version check
	err = xRefTable.ValidateVersion(fmt.Sprintf("dict=%s entry=%s", dictName, entryName), sinceVersion)
	if err != nil {
		return err
	}

	switch obj.(type) {

	case types.PDFInteger, types.PDFDict:
		// no further processing

	default:
		return errors.Errorf("validateIntOrDictEntry: dict=%s entry=%s invalid type", dictName, entryName)
	}

	logInfoValidate.Printf("validateIntOrDictEntry end: entry=%s\n", entryName)

	return nil
}

func validateBooleanOrStreamEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName, entryName string, required bool, sinceVersion types.PDFVersion) error {

	logInfoValidate.Printf("validateBooleanOrStreamEntry begin: entry=%s\n", entryName)

	obj, err := dict.Entry(dictName, entryName, required)
	if err != nil || obj == nil {
		return err
	}

	obj, err = xRefTable.Dereference(obj)
	if err != nil {
		return err
	}
	if obj == nil {
		if required {
			return errors.Errorf("validateBooleanOrStreamEntry: dict=%s required entry=%s is nil", dictName, entryName)
		}
		logInfoValidate.Printf("validateBooleanOrStreamEntry end: optional entry %s is nil\n", entryName)
		return nil
	}

	// Version check
	err = xRefTable.ValidateVersion(fmt.Sprintf("dict=%s entry=%s", dictName, entryName), sinceVersion)
	if err != nil {
		return err
	}

	switch obj.(type) {

	case types.PDFBoolean, types.PDFStreamDict:
		// no further processing

	default:
		return errors.Errorf("validateBooleanOrStreamEntry: dict=%s entry=%s invalid type", dictName, entryName)
	}

	logInfoValidate.Printf("validateBooleanOrStreamEntry end: entry=%s\n", entryName)

	return nil
}

// TODO move to 3D annotation.
func validateStreamDictOrDictEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName, entryName string, required bool, sinceVersion types.PDFVersion) error {

	logInfoValidate.Printf("validateStreamDictOrDictEntry begin: entry=%s\n", entryName)

	obj, err := dict.Entry(dictName, entryName, required)
	if err != nil || obj == nil {
		return err
	}

	obj, err = xRefTable.Dereference(obj)
	if err != nil {
		return err
	}
	if obj == nil {
		if required {
			return errors.Errorf("validateStreamDictOrDictEntry: dict=%s required entry=%s is nil", dictName, entryName)
		}
		logInfoValidate.Printf("validateStreamDictOrDictEntry end: optional entry %s is nil\n", entryName)
		return nil
	}

	// Version check
	err = xRefTable.ValidateVersion(fmt.Sprintf("dict=%s entry=%s", dictName, entryName), sinceVersion)
	if err != nil {
		return err
	}

	switch obj.(type) {

	case types.PDFStreamDict:
		// TODO validate 3D stream dict

	case types.PDFDict:
		// TODO validate 3D reference dict

	default:
		return errors.Errorf("validateStreamDictOrDictEntry: dict=%s entry=%s invalid type", dictName, entryName)
	}

	logInfoValidate.Printf("validateStreamDictOrDictEntry end: entry=%s\n", entryName)

	return nil
}

func validateIntegerOrArrayOfIntegerEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName, entryName string, required bool, sinceVersion types.PDFVersion) error {

	logInfoValidate.Printf("validateIntegerOrArrayOfIntegerEntry begin: entry=%s\n", entryName)

	obj, err := dict.Entry(dictName, entryName, required)
	if err != nil || obj == nil {
		return err
	}

	obj, err = xRefTable.Dereference(obj)
	if err != nil {
		return err
	}
	if obj == nil {
		if required {
			return errors.Errorf("validateIntegerOrArrayOfIntegerEntry: dict=%s required entry=%s is nil", dictName, entryName)
		}
		logInfoValidate.Printf("validateIntegerOrArrayOfIntegerEntry end: optional entry %s is nil\n", entryName)
		return nil
	}

	// Version check
	err = xRefTable.ValidateVersion(fmt.Sprintf("dict=%s entry=%s", dictName, entryName), sinceVersion)
	if err != nil {
		return err
	}

	switch obj := obj.(type) {

	case types.PDFInteger:
		// no further processing

	case types.PDFArray:

		for i, obj := range obj {

			obj, err = xRefTable.Dereference(obj)
			if err != nil {
				return err
			}

			if obj == nil {
				continue
			}

			_, ok := obj.(types.PDFInteger)
			if !ok {
				return errors.Errorf("validateIntegerOrArrayOfIntegerEntry: dict=%s entry=%s invalid type at index %d\n", dictName, entryName, i)
			}

		}

	default:
		return errors.Errorf("validateIntegerOrArrayOfIntegerEntry: dict=%s entry=%s invalid type", dictName, entryName)
	}

	logInfoValidate.Printf("validateIntegerOrArrayOfIntegerEntry end: entry=%s\n", entryName)

	return nil
}

func validateNameOrArrayOfNameEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName, entryName string, required bool, sinceVersion types.PDFVersion) error {

	logInfoValidate.Printf("validateNameOrArrayOfNameEntry begin: entry=%s\n", entryName)

	obj, err := dict.Entry(dictName, entryName, required)
	if err != nil || obj == nil {
		return err
	}

	obj, err = xRefTable.Dereference(obj)
	if err != nil {
		return err
	}
	if obj == nil {
		if required {
			return errors.Errorf("validateNameOrArrayOfNameEntry: dict=%s required entry=%s is nil", dictName, entryName)
		}
		logInfoValidate.Printf("validateNameOrArrayOfNameEntry end: optional entry %s is nil\n", entryName)
		return nil
	}

	// Version check
	err = xRefTable.ValidateVersion(fmt.Sprintf("dict=%s entry=%s", dictName, entryName), sinceVersion)
	if err != nil {
		return err
	}

	switch obj := obj.(type) {

	case types.PDFName:
		// no further processing

	case types.PDFArray:

		for i, obj := range obj {

			obj, err = xRefTable.Dereference(obj)
			if err != nil {
				return err
			}

			if obj == nil {
				continue
			}

			_, ok := obj.(types.PDFName)
			if !ok {
				err = errors.Errorf("validateNameOrArrayOfNameEntry: dict=%s entry=%s invalid type at index %d\n", dictName, entryName, i)
				return err
			}

		}

	default:
		return errors.Errorf("validateNameOrArrayOfNameEntry: dict=%s entry=%s invalid type", dictName, entryName)
	}

	logInfoValidate.Printf("validateNameOrArrayOfNameEntry end: entry=%s\n", entryName)

	return nil
}

func validateBooleanOrArrayOfBooleanEntry(xRefTable *types.XRefTable, dict *types.PDFDict, dictName, entryName string, required bool, sinceVersion types.PDFVersion) error {

	logInfoValidate.Printf("validateBooleanOrArrayOfBooleanEntry begin: entry=%s\n", entryName)

	obj, err := dict.Entry(dictName, entryName, required)
	if err != nil || obj == nil {
		return err
	}

	obj, err = xRefTable.Dereference(obj)
	if err != nil {
		return err
	}
	if obj == nil {
		if required {
			return errors.Errorf("validateBooleanOrArrayOfBooleanEntry: dict=%s required entry=%s is nil", dictName, entryName)
		}
		logInfoValidate.Printf("validateBooleanOrArrayOfBooleanEntry end: optional entry %s is nil\n", entryName)
		return nil
	}

	// Version check
	err = xRefTable.ValidateVersion(fmt.Sprintf("dict=%s entry=%s", dictName, entryName), sinceVersion)
	if err != nil {
		return err
	}

	switch obj := obj.(type) {

	case types.PDFBoolean:
		// no further processing

	case types.PDFArray:

		for i, obj := range obj {

			obj, err = xRefTable.Dereference(obj)
			if err != nil {
				return err
			}

			if obj == nil {
				continue
			}

			_, ok := obj.(types.PDFBoolean)
			if !ok {
				return errors.Errorf("validateBooleanOrArrayOfBooleanEntry: dict=%s entry=%s invalid type at index %d\n", dictName, entryName, i)
			}

		}

	default:
		return errors.Errorf("validateBooleanOrArrayOfBooleanEntry: dict=%s entry=%s invalid type", dictName, entryName)
	}

	logInfoValidate.Printf("validateBooleanOrArrayOfBooleanEntry end: entry=%s\n", entryName)

	return nil
}
