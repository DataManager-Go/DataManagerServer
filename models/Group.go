package models

import (
	"github.com/jinzhu/gorm"
)

//Group a group in DB
type Group struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Namespace   *Namespace
	NamespaceID int64 `sql:"index" gorm:"not null"`
}

//Insert inserts group into DB
func (group *Group) Insert(db *gorm.DB) error {
	//Use default namespace if nil
	group.Namespace = group.GetNamespace()
	return db.Create(group).Error
}

//GetNamespace return namespace of group
func (group Group) GetNamespace() *Namespace {
	if group.Namespace == nil {
		return &DefaultNamespace
	}

	return group.Namespace
}

//GroupsFromStringArr return tag array from string array
func GroupsFromStringArr(arr []string, namespace Namespace) []Group {
	var tags []Group
	for _, tag := range arr {
		tags = append(tags, Group{
			Name:      tag,
			Namespace: &namespace,
		})
	}
	return tags
}

//FindGroups find group in db
func FindGroups(db *gorm.DB, sGroups []string, namespace *Namespace) []Group {
	var groups []Group
	db.Model(&Group{}).Where("name in (?) AND namespace_id = ?", sGroups, namespace.ID).Find(&groups)
	return groups
}
