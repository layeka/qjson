package qjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/layeka/qutils"
)

const (
	JSONStringEmpty = ""
	JSONStringNull  = "null"
)

type jsonArray []interface{}

func (this *jsonArray) Add(val interface{}) *jsonArray {
	*this = append(*this, val)
	return this
}

func (this *jsonArray) Get(index int) (val interface{}, b bool) {
	if len(*this) > index {
		return (*this)[index], true
	}
	return nil, false
}
func (this *jsonArray) Last() (val interface{}, b bool) {
	l := len(*this)
	if l > 0 {
		return (*this)[l-1], true
	}
	return nil, false
}

func NewjsonArray(data []interface{}) *jsonArray {
	arr := &jsonArray{}
	for _, item := range data {
		switch val := item.(type) {
		case []interface{}:
			arr.Add(NewjsonArray(val))
		case *[]interface{}:
			arr.Add(NewjsonArray(*val))
		case map[string]interface{}:
			arr.Add(NewjsonObject(val))
		case *map[string]interface{}:
			arr.Add(NewjsonObject(*val))
		default:
			arr.Add(val)
		}
	}
	return arr
}

type jsonObject map[string]interface{}

func (this *jsonObject) Add(key string, val interface{}) *jsonObject {
	(*this)[key] = val
	return this
}
func (this *jsonObject) Get(key string) (val interface{}, b bool) {
	val, b = (*this)[key]
	return
}

func NewjsonObject(data map[string]interface{}) *jsonObject {
	obj := &jsonObject{}
	for key, item := range data {
		switch val := item.(type) {
		case []interface{}:
			obj.Add(key, NewjsonArray(val))
		case *[]interface{}:
			obj.Add(key, NewjsonArray(*val))
		case map[string]interface{}:
			obj.Add(key, NewjsonObject(val))
		case *map[string]interface{}:
			obj.Add(key, NewjsonObject(*val))
		default:
			obj.Add(key, val)
		}
	}
	return obj
}

type QJson struct {
	data interface{}
}

func NewJSON(body []byte) (*QJson, error) {
	j := new(QJson)
	err := j.UnmarshalJSON(body)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func NewObjectJSON() *QJson {
	return &QJson{&jsonObject{}}
}
func NewArrayJSON() *QJson {
	return &QJson{&jsonArray{}}
}

// Implements the json.Marshaler interface.
func (this *QJson) MarshalJSON() ([]byte, error) {
	return json.Marshal(&this.data)
}

// Implements the json.Unmarshaler interface.
func (this *QJson) UnmarshalJSON(p []byte) error {
	dec := json.NewDecoder(bytes.NewBuffer(p))
	dec.UseNumber()
	var obj interface{}
	err := dec.Decode(&obj)
	if err != nil {
		return err
	}
	switch val := obj.(type) {
	case []interface{}:
		this.data = NewjsonArray(val)
	case map[string]interface{}:
		this.data = NewjsonObject(val)
	default:
		this.data = obj
	}
	return nil
}
func (this *QJson) IsArray() bool {
	_, ok := this.data.(*jsonArray)
	return ok
}
func (this *QJson) asArray() (*jsonArray, error) {
	val, ok := this.data.(*jsonArray)
	if ok {
		return val, nil
	}
	return nil, errors.New("type assertion to *jsonArray failed")
}
func (this *QJson) IsObject() bool {
	_, ok := this.data.(*jsonObject)
	return ok
}
func (this *QJson) asObject() (*jsonObject, error) {
	val, ok := this.data.(*jsonObject)
	if ok {
		return val, nil
	}
	return nil, errors.New("type assertion to *jsonObject failed")
}

func (this *QJson) ArrayAdd(obj interface{}) *QJson {
	arr, err := this.asArray()
	if err == nil {
		switch val := obj.(type) {
		case []interface{}:
			arr.Add(NewjsonArray(val))
		case *[]interface{}:
			arr.Add(NewjsonArray(*val))
		case map[string]interface{}:
			arr.Add(NewjsonObject(val))
		case *map[string]interface{}:
			arr.Add(NewjsonObject(*val))
		default:
			arr.Add(val)
		}
	}
	return this
}
func (this *QJson) ArrayGet(index int) *QJson {
	arr, err := this.asArray()
	if err == nil {
		val, ok := arr.Get(index)
		if ok {
			return &QJson{val}
		}
	}
	return &QJson{nil}
}
func (this *QJson) ArrayNewArray() *QJson {
	arr, err := this.asArray()
	if err == nil {
		val, b := arr.Add(&jsonArray{}).Last()
		if b {
			return &QJson{val}
		}
	}
	return &QJson{nil}
}

func (this *QJson) ArrayNewObject() *QJson {
	arr, err := this.asArray()
	if err == nil {
		val, b := arr.Add(&jsonObject{}).Last()
		if b {
			return &QJson{val}
		}
	}
	return &QJson{nil}
}
func (this *QJson) Exists(key string) bool {
	jobj, err := this.asObject()
	if err != nil {
		return false
	}
	_, ok := jobj.Get(key)
	return ok
}
func (this *QJson) ObjectAdd(key string, obj interface{}) *QJson {
	jobj, err := this.asObject()
	if err == nil {
		switch val := obj.(type) {
		case []interface{}:
			jobj.Add(key, NewjsonArray(val))
		case *[]interface{}:
			jobj.Add(key, NewjsonArray(*val))
		case map[string]interface{}:
			jobj.Add(key, NewjsonObject(val))
		case *map[string]interface{}:
			jobj.Add(key, NewjsonObject(*val))
		default:
			jobj.Add(key, val)
		}
	}
	return this
}
func (this *QJson) ObjectGet(key string) *QJson {
	obj, err := this.asObject()
	if err == nil {
		val, ok := obj.Get(key)
		if ok {
			return &QJson{val}
		}
	}
	return &QJson{nil}
}
func (this *QJson) ObjectNewObject(key string) *QJson {
	return this.ObjectAdd(key, &jsonObject{}).ObjectGet(key)
}

func (this *QJson) ObjectNewArray(key string) *QJson {
	return this.ObjectAdd(key, &jsonArray{}).ObjectGet(key)
}

func (this *QJson) String() string {
	b, err := json.Marshal(this.data)
	if err != nil {
		return ""
	}
	return string(b)
}
func (this *QJson) MustBool(args ...bool) bool {
	val, err := convertBool(this.data)
	if err != nil {
		return qutils.ArgsBool(args).Default(0)
	}
	return val
}
func (this *QJson) MustInt(args ...int64) int64 {
	val, err := convertInt64(this.data)
	if err != nil {
		fmt.Println(reflect.TypeOf(this.data))
		return qutils.ArgsInt64(args).Default(0)
	}
	return val
}

func (this *QJson) MustFloat(args ...float64) float64 {
	val, err := convertFloat64(this.data)
	if err != nil {
		return qutils.ArgsFloat64(args).Default(0)
	}
	return val
}

func (this *QJson) MustString(args ...string) string {
	val, err := convertString(this.data)
	if err != nil {
		return qutils.ArgsString(args).Default(0)
	}
	return val
}

func (this *QJson) MustArray(args ...[]interface{}) []interface{} {
	val, err := this.asArray()
	if err != nil {
		return qutils.ArgsArray(args).Default(0)
	}
	return ([]interface{})(*val)
}

func (this *QJson) MustObject(args ...map[string]interface{}) map[string]interface{} {
	val, err := this.asObject()
	if err != nil {
		return qutils.ArgsObject(args).Default(0)
	}
	return (map[string]interface{})(*val)
}

func convertBool(o interface{}) (bool, error) {
	switch val := o.(type) {
	case nil:
		return false, nil
	case bool:
		return val, nil
	case string:
		return strconv.ParseBool(val)
	case json.Number:
		{
			v, err := val.Int64()
			if err == nil {
				return v != 0, nil
			}
		}
	case int, int8, int16, int32, int64:
		return qutils.SafeInt64(val) != 0, nil
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return qutils.SafeUInt64(val) != 0, nil
	case float32, float64:
		return qutils.SafeFloat64(val) != 0, nil
	}
	return false, errors.New("The json can not be converted to a Boolean value")
}

func convertString(o interface{}) (string, error) {
	switch val := o.(type) {
	case nil:
		return JSONStringNull, nil
	case bool:
		return strconv.FormatBool(val), nil
	case string:
		return val, nil
	case json.Number:
		return val.String(), nil
	case int, int8, int16, int32, int64:
		return strconv.FormatInt(qutils.SafeInt64(val), 10), nil
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return strconv.FormatUint(qutils.SafeUInt64(val), 10), nil
	case float32, float64:
		return strconv.FormatFloat(qutils.SafeFloat64(val), 'f', 18, 64), nil
	default:
		{
			s, ok := o.(string)
			if ok {
				return s, nil
			}
		}
	}
	return JSONStringEmpty, errors.New("The json can not be converted to a string value")
}

func convertInt64(o interface{}) (int64, error) {
	switch val := o.(type) {
	case nil:
		return 0, nil
	case bool:
		{
			if val {
				return 1, nil
			}
			return 0, nil
		}
	case string:
		return strconv.ParseInt(val, 10, 0)
	case json.Number:
		return val.Int64()
	case int, int8, int16, int32, int64:
		return qutils.SafeInt64(val), nil
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return int64(qutils.SafeUInt64(val)), nil
	case float32, float64:
		return int64(qutils.SafeFloat64(val)), nil
	}
	return 0, errors.New("The json can not be converted to a int64 value")
}
func convertFloat64(o interface{}) (float64, error) {
	switch val := o.(type) {
	case nil:
		return 0, nil
	case bool:
		{
			if val {
				return 1, nil
			}
			return 0, nil
		}
	case string:
		return strconv.ParseFloat(val, 64)
	case json.Number:
		return val.Float64()
	case int, int8, int16, int32, int64:
		return float64(qutils.SafeInt64(val)), nil
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return float64(qutils.SafeUInt64(val)), nil
	case float32, float64:
		return qutils.SafeFloat64(val), nil
	}
	return 0, errors.New("The json can not be converted to a float64 value")
}
