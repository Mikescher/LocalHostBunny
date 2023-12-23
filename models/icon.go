package models

import (
	"gogs.mikescher.com/BlackForestBytes/goext/rfctime"
)

type Icon struct {
	IconID      IconID                  `bson:"_id,omitempty" json:"id"`
	Checksum    string                  `bson:"checksum"      json:"checksum"`
	Data        []byte                  `bson:"data"          json:"data"`
	ContentType string                  `bson:"contentType"   json:"contentType"`
	Time        rfctime.RFC3339NanoTime `bson:"time"          json:"time"`
}
