package models

import (
	"github.com/jinzhu/gorm"
)

//Group a group in DB
type Group struct {
	gorm.Model
	Name        string     `gorm:"not null"`
	NamespaceID uint       `sql:"index" gorm:"not null"`
	Namespace   *Namespace `gorm:"association_autoupdate:false;association_autocreate:false"`
	UserID      uint       `sql:"index" gorm:"not null"`
	User        *User      `gorm:"association_autoupdate:false;association_autocreate:false"`
}

//Insert inserts group into DB
func (group *Group) Insert(db *gorm.DB, user *User) error {
	//Use default namespace if nil
	group.Namespace = group.GetNamespace()
	group.User = user
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
func GroupsFromStringArr(arr []string, namespace Namespace, user *User) []Group {
	var tags []Group
	for _, tag := range arr {
		tags = append(tags, Group{
			Name:      tag,
			User:      user,
			UserID:    user.ID,
			Namespace: &namespace,
		})
	}
	return tags
}

//GroupArrToStringArr return string arr from group
func GroupArrToStringArr(groups []Group) []string {
	var str []string
	for _, group := range groups {
		str = append(str, group.Name)
	}
	return str
}

//FindGroups find group in db
func FindGroups(db *gorm.DB, sGroups []string, namespace *Namespace) []Group {
	var groups []Group
	db.Model(&Group{}).Where("name in (?) AND namespace_id = ?", sGroups, namespace.ID).Find(&groups)
	return groups
}

//GetGroup returns or creates a tag
func GetGroup(db *gorm.DB, name string, namespace *Namespace, user *User) *Group {
	var group Group
	db.Where(&Group{
		Name:        name,
		NamespaceID: namespace.ID,
		UserID:      user.ID,
	}).FirstOrCreate(&group)

	return &group
}

//FindGroup finds a tag
func FindGroup(db *gorm.DB, name string, namespace *Namespace, user *User) (*Group, error) {
	var group Group

	err := db.Where(&Group{
		Name:        name,
		NamespaceID: namespace.ID,
		UserID:      user.ID,
	}).Find(&group).Error

	if err != nil {
		return nil, err
	}

	return &group, nil
}
