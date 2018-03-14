package model

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/qb0C80aE/clay/extension"
	"net/url"
	"reflect"
	"sort"
	"sync"
)

var typeNameUserDefinedModelDefinitionMap = map[string]*UserDefinedModelDefinition{}
var typeNameUserDefinedModelDefinitionMapMutex = new(sync.Mutex)

var typeNameTypeMap = map[string]reflect.Type{
	"int":     reflect.TypeOf(int(0)),
	"int8":    reflect.TypeOf(int(0)),
	"int16":   reflect.TypeOf(int(0)),
	"int32":   reflect.TypeOf(int(0)),
	"int64":   reflect.TypeOf(int64(0)),
	"uint":    reflect.TypeOf(uint(0)),
	"uint8":   reflect.TypeOf(uint(0)),
	"uint16":  reflect.TypeOf(uint(0)),
	"uint32":  reflect.TypeOf(uint(0)),
	"uint64":  reflect.TypeOf(uint64(0)),
	"float":   reflect.TypeOf(float32(0)),
	"float64": reflect.TypeOf(float64(0)),
	"bool":    reflect.TypeOf(bool(false)),
	"string":  reflect.TypeOf(string("")),
}

// UserDefinedModelDefinition is the model class what represents raw template
type UserDefinedModelDefinition struct {
	Base
	TypeName            string                             `json:"type_name" binding:"required" clay:"key_parameter" `
	ResourceName        string                             `json:"resource_name" binding:"required"`
	ToBeMigrated        bool                               `json:"to_be_migrated" binding:"required"`
	IsControllerEnabled bool                               `json:"is_controller_enabled" binding:"required"`
	SQLBeforeMigration  string                             `json:"sql_before_migration"`
	SQLAfterMigration   string                             `json:"sql_after_migration"`
	SQLDesignExtraction string                             `json:"sql_design_extraction"`
	SQLDesignDeletion   string                             `json:"sql_design_deletion"`
	Fields              []*UserDefinedModelFieldDefinition `json:"fields"`
	structFieldList     []reflect.StructField
}

// NewUserDefinedModelDefinition creates a template raw model instance
func NewUserDefinedModelDefinition() *UserDefinedModelDefinition {
	return &UserDefinedModelDefinition{}
}

// GetContainerForMigration returns its container for migration, if no need to be migrated, just return null
func (receiver *UserDefinedModelDefinition) GetContainerForMigration() (interface{}, error) {
	return nil, nil
}

// GetSingle corresponds HTTP GET message and handles a request for a single resource to get the information
func (receiver *UserDefinedModelDefinition) GetSingle(model extension.Model, db *gorm.DB, parameters gin.Params, urlValues url.Values, queryFields string) (interface{}, error) {
	result, exists := typeNameUserDefinedModelDefinitionMap[parameters.ByName("type_name")]

	if !exists {
		return nil, errors.New("record not found")
	}

	return result, nil
}

// GetMulti corresponds HTTP GET message and handles a request for multi resource to get the list of information
func (receiver *UserDefinedModelDefinition) GetMulti(model extension.Model, db *gorm.DB, parameters gin.Params, urlValues url.Values, queryFields string) (interface{}, error) {
	typeNameUserDefinedModelDefinitionMapMutex.Lock()
	defer typeNameUserDefinedModelDefinitionMapMutex.Unlock()

	keyList := make([]string, 0, len(typeNameUserDefinedModelDefinitionMap))
	for key := range typeNameUserDefinedModelDefinitionMap {
		keyList = append(keyList, key)
	}

	sort.Strings(keyList)

	userDefinedModelDefinitionList := make([]*UserDefinedModelDefinition, len(typeNameUserDefinedModelDefinitionMap))

	for i, key := range keyList {
		userDefinedModelDefinitionList[i] = typeNameUserDefinedModelDefinitionMap[key]
	}

	return userDefinedModelDefinitionList, nil
}

// Create corresponds HTTP POST message and handles a request for multi resource to create a new information
func (receiver *UserDefinedModelDefinition) Create(model extension.Model, db *gorm.DB, _ gin.Params, _ url.Values, inputContainer interface{}) (interface{}, error) {
	typeNameUserDefinedModelDefinitionMapMutex.Lock()
	defer typeNameUserDefinedModelDefinitionMapMutex.Unlock()

	userDefinedModelDefinition := NewUserDefinedModelDefinition()
	if err := extension.ConvertInputMapToContainer(inputContainer, userDefinedModelDefinition); err != nil {
		return nil, err
	}

	newUserDefinedModel := NewUserDefinedModel()

	structFieldList := []reflect.StructField{}
	for _, field := range userDefinedModelDefinition.Fields {
		reflectType, exists := typeNameTypeMap[field.TypeName]
		if !exists {
			if field.IsSlice {
				reflectType = extension.NewStructFieldTypeProxy(field.TypeName, reflect.Slice)
			} else {
				reflectType = extension.NewStructFieldTypeProxy(field.TypeName, reflect.Struct)
			}
		}

		structField := reflect.StructField{
			Name: field.Name,
			Tag:  reflect.StructTag(field.Tag),
			Type: reflectType,
		}
		structFieldList = append(structFieldList, structField)
	}
	newUserDefinedModel.typeName = userDefinedModelDefinition.TypeName
	newUserDefinedModel.resourceName = userDefinedModelDefinition.ResourceName
	newUserDefinedModel.toBeMigrated = userDefinedModelDefinition.ToBeMigrated
	newUserDefinedModel.isControllerEnabled = userDefinedModelDefinition.IsControllerEnabled
	newUserDefinedModel.sqlBeforeMigration = userDefinedModelDefinition.SQLBeforeMigration
	newUserDefinedModel.sqlAfterMigration = userDefinedModelDefinition.SQLAfterMigration
	newUserDefinedModel.sqlDesignExtraction = userDefinedModelDefinition.SQLDesignExtraction
	newUserDefinedModel.sqlDesignDeletion = userDefinedModelDefinition.SQLDesignDeletion
	newUserDefinedModel.structFieldList = structFieldList

	extension.RegisterModel(newUserDefinedModel)
	extension.RegisterDesignAccessor(newUserDefinedModel)

	initializerList := []extension.Initializer{newUserDefinedModel}
	modelList := []extension.Model{newUserDefinedModel}

	if _, err := extension.SetupModel(db, initializerList, modelList); err != nil {
		return nil, err
	}

	typeNameUserDefinedModelDefinitionMap[userDefinedModelDefinition.TypeName] = userDefinedModelDefinition

	return userDefinedModelDefinition, nil
}

// GetTotal returns the count of for multi resource
func (receiver *UserDefinedModelDefinition) GetTotal(_ extension.Model, _ *gorm.DB) (int, error) {
	return len(typeNameUserDefinedModelDefinitionMap), nil
}

func init() {
	extension.RegisterModel(NewUserDefinedModelDefinition())
}