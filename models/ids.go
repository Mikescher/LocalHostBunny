package models

import (
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//go:generate go run ../_gen/id-generate.go -- ids_gen.go

type EntityID interface {
	ObjID() (primitive.ObjectID, error)
	String() string
	MarshalBSONValue() (bsontype.Type, []byte, error)
	AsAny() AnyID
}

type AnyID string //@id:type

type JobLogID string //@id:type

type JobExecutionID string //@id:type
