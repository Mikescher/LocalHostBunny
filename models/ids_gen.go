// Code generated by id-generate.go DO NOT EDIT.

package models

import "go.mongodb.org/mongo-driver/bson"
import "go.mongodb.org/mongo-driver/bson/bsontype"
import "go.mongodb.org/mongo-driver/bson/primitive"
import "gogs.mikescher.com/BlackForestBytes/goext/exerr"

const ChecksumIDGenerator = "8fa696914bf8d1c1c4b9f80be45d0a9dfbd0fed789856bf84ced2979c789b958" // GoExtVersion: 0.0.288

// ================================ AnyID (ids.go) ================================

func (i AnyID) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if objId, err := primitive.ObjectIDFromHex(string(i)); err == nil {
		return bson.MarshalValue(objId)
	} else {
		return 0, nil, exerr.New(exerr.TypeMarshalEntityID, "Failed to marshal AnyID("+i.String()+") to ObjectId").Str("value", string(i)).Type("type", i).Build()
	}
}

func (i AnyID) String() string {
	return string(i)
}

func (i AnyID) ObjID() (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(string(i))
}

func (i AnyID) Valid() bool {
	_, err := primitive.ObjectIDFromHex(string(i))
	return err == nil
}

func (i AnyID) AsAny() AnyID {
	return AnyID(i)
}

func NewAnyID() AnyID {
	return AnyID(primitive.NewObjectID().Hex())
}

// ================================ JobLogID (ids.go) ================================

func (i JobLogID) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if objId, err := primitive.ObjectIDFromHex(string(i)); err == nil {
		return bson.MarshalValue(objId)
	} else {
		return 0, nil, exerr.New(exerr.TypeMarshalEntityID, "Failed to marshal JobLogID("+i.String()+") to ObjectId").Str("value", string(i)).Type("type", i).Build()
	}
}

func (i JobLogID) String() string {
	return string(i)
}

func (i JobLogID) ObjID() (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(string(i))
}

func (i JobLogID) Valid() bool {
	_, err := primitive.ObjectIDFromHex(string(i))
	return err == nil
}

func (i JobLogID) AsAny() AnyID {
	return AnyID(i)
}

func NewJobLogID() JobLogID {
	return JobLogID(primitive.NewObjectID().Hex())
}

// ================================ JobExecutionID (ids.go) ================================

func (i JobExecutionID) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if objId, err := primitive.ObjectIDFromHex(string(i)); err == nil {
		return bson.MarshalValue(objId)
	} else {
		return 0, nil, exerr.New(exerr.TypeMarshalEntityID, "Failed to marshal JobExecutionID("+i.String()+") to ObjectId").Str("value", string(i)).Type("type", i).Build()
	}
}

func (i JobExecutionID) String() string {
	return string(i)
}

func (i JobExecutionID) ObjID() (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(string(i))
}

func (i JobExecutionID) Valid() bool {
	_, err := primitive.ObjectIDFromHex(string(i))
	return err == nil
}

func (i JobExecutionID) AsAny() AnyID {
	return AnyID(i)
}

func NewJobExecutionID() JobExecutionID {
	return JobExecutionID(primitive.NewObjectID().Hex())
}