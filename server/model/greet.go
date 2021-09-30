/**
 *  MindLab
 *
 *  Create by songli on 2021/09/30
 *  Copyright © 2021 imind.tech All rights reserved.
 */

package model

import (
	"gorm.io/gorm"
	"reflect"
	"time"
)

type Greet struct {
	Id             int32  `gorm:"primary_key" redis:"id"`
	Name           string `redis:"name,omitempty"`
	ViewNum        int32  `redis:"view_num,omitempty"`
	Status         int32  `redis:"status,omitempty"`
	CreateTime     int64  `redis:"create_time,omitempty"`
	CreateDatetime string `redis:"create_datetime,omitempty"`
	UpdateDatetime string `redis:"update_datetime,omitempty"`
}

func (Greet) TableName() string {
	return "tbl_greet"
}

func (m *Greet) BeforeCreate(tx *gorm.DB) error {
	m.CreateDatetime = time.Now().Format("2006-01-02 15:04:05")
	m.UpdateDatetime = time.Now().Format("2006-01-02 15:04:05")
	return nil
}

func (m *Greet) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateDatetime = time.Now().Format("2006-01-02 15:04:05")
	return nil
}

func (m Greet) IsEmpty() bool {
	return reflect.DeepEqual(m, Greet{})
}
