package logics

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/qb0C80aE/clay/extensions"
	"github.com/qb0C80aE/clay/models"
	"github.com/qb0C80aE/clay/utils/mapstruct"
	"net/url"
	"strconv"
)

type templatePersistentParameterLogic struct {
	*BaseLogic
}

func newTemplatePersistentParameterLogic() *templatePersistentParameterLogic {
	logic := &templatePersistentParameterLogic{
		BaseLogic: NewBaseLogic(
			models.SharedTemplatePersistentParameterModel(),
		),
	}
	return logic
}

func (logic *templatePersistentParameterLogic) GetSingle(db *gorm.DB, parameters gin.Params, _ url.Values, queryFields string) (interface{}, error) {

	templatePersistentParameter := &models.TemplatePersistentParameter{}

	if err := db.Select(queryFields).First(templatePersistentParameter, parameters.ByName("id")).Error; err != nil {
		return nil, err
	}

	return templatePersistentParameter, nil

}

func (logic *templatePersistentParameterLogic) GetMulti(db *gorm.DB, _ gin.Params, _ url.Values, queryFields string) (interface{}, error) {
	templatePersistentParameters := []*models.TemplatePersistentParameter{}

	if err := db.Select(queryFields).Find(&templatePersistentParameters).Error; err != nil {
		return nil, err
	}

	return templatePersistentParameters, nil
}

func (logic *templatePersistentParameterLogic) Create(db *gorm.DB, _ gin.Params, _ url.Values, data interface{}) (interface{}, error) {

	templatePersistentParameter := data.(*models.TemplatePersistentParameter)

	if err := db.Create(templatePersistentParameter).Error; err != nil {
		return nil, err
	}

	return templatePersistentParameter, nil

}

func (logic *templatePersistentParameterLogic) Update(db *gorm.DB, parameters gin.Params, _ url.Values, data interface{}) (interface{}, error) {

	templatePersistentParameter := data.(*models.TemplatePersistentParameter)
	templatePersistentParameter.ID, _ = strconv.Atoi(parameters.ByName("id"))

	if err := db.Save(templatePersistentParameter).Error; err != nil {
		return nil, err
	}

	return templatePersistentParameter, nil

}

func (logic *templatePersistentParameterLogic) Delete(db *gorm.DB, parameters gin.Params, _ url.Values) error {

	templatePersistentParameter := &models.TemplatePersistentParameter{}

	if err := db.First(templatePersistentParameter, parameters.ByName("id")).Error; err != nil {
		return err
	}

	return db.Delete(templatePersistentParameter).Error
}

func (logic *templatePersistentParameterLogic) ExtractFromDesign(db *gorm.DB) (string, interface{}, error) {
	templatePersistentParameters := []*models.TemplatePersistentParameter{}
	if err := db.Select("*").Find(&templatePersistentParameters).Error; err != nil {
		return "", nil, err
	}
	return extensions.RegisteredResourceName(models.SharedTemplatePersistentParameterModel()), templatePersistentParameters, nil
}

func (logic *templatePersistentParameterLogic) DeleteFromDesign(db *gorm.DB) error {
	return db.Delete(models.SharedTemplatePersistentParameterModel()).Error
}

func (logic *templatePersistentParameterLogic) LoadToDesign(db *gorm.DB, data interface{}) error {
	container := []*models.TemplatePersistentParameter{}
	design := data.(*models.Design)
	if value, exists := design.Content[extensions.RegisteredResourceName(models.SharedTemplatePersistentParameterModel())]; exists {
		if err := mapstruct.MapToStruct(value.([]interface{}), &container); err != nil {
			return err
		}
		for _, templatePersistentParameter := range container {
			if err := db.Create(templatePersistentParameter).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

var uniqueTemplatePersistentParameterLogic = newTemplatePersistentParameterLogic()

// UniqueTemplatePersistentParameterLogic returns the unique template persistent parameter logic instance
func UniqueTemplatePersistentParameterLogic() extensions.Logic {
	return uniqueTemplatePersistentParameterLogic
}

func init() {
	extensions.RegisterDesignAccessor(uniqueTemplatePersistentParameterLogic)
	extensions.RegisterTemplateParameterGenerator(models.SharedTemplatePersistentParameterModel(), uniqueTemplatePersistentParameterLogic)
}
