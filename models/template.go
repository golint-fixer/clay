package models

import (
	"github.com/qb0C80aE/clay/extension"
	"reflect"
)

type TemplateExternalParameter struct {
	ID         int    `json:"id" gorm:"primary_key;AUTO_INCREMENT"`
	TemplateID int    `json:"template_id" gorm:"index" sql:"type:integer references templates(id) on delete cascade"`
	Name       string `json:"name"`
	Value      string `json:"value"`
}

type Template struct {
	ID                         int                          `json:"id" form:"id" gorm:"primary_key;AUTO_INCREMENT"`
	Name                       string                       `json:"name" form:"name"`
	TemplateContent            string                       `json:"template_content" form:"template_content"`
	TemplateExternalParameters []*TemplateExternalParameter `json:"template_external_parameters"`
}

func init() {
	extension.RegisterModelType(reflect.TypeOf(TemplateExternalParameter{}))
	extension.RegisterModelType(reflect.TypeOf(Template{}))
}

var TemplateExternalParameterModel = &TemplateExternalParameter{}
var TemplateModel = &Template{}