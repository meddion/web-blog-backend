package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const collNameStatic = "files"

// File struct is a model for a files collection
type File struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Dir          string             `bson:"dir"`
	Name         string             `bson:"name"`
	Ext          string             `bson:"ext"`
	File         primitive.Binary   `bson:"file,omitempty"`
	CreationTime primitive.DateTime `bson:"creation_time,omitempty"`
}

func NewEmptyFile(dir, name, ext string) *File {
	return &File{
		Dir:          dir,
		Name:         name,
		Ext:          ext,
		CreationTime: primitive.NewDateTimeFromTime(time.Now()),
	}
}

func NewFile(dir, name, ext string, data []byte) *File {
	file := NewEmptyFile(dir, name, ext)
	file.File.Data = data
	return file
}

func (s *File) getFilter() bson.M {
	return bson.M{"dir": s.Dir, "name": s.Name, "ext": s.Ext}
}

// Save saves the file in DB
func (s *File) Save(ctx context.Context) (*mongo.UpdateResult, error) {
	update := bson.M{"$set": s}
	opts := options.Update().SetUpsert(true)

	result, err := GetDB().Collection(collNameStatic).UpdateOne(ctx, s.getFilter(), update, opts)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Get fetches a binary of the file
func (s *File) Get(ctx context.Context) error {
	opts := options.FindOne().SetProjection(bson.D{{"file", 1}})

	if err := GetDB().Collection(collNameStatic).FindOne(ctx, s.getFilter(), opts).Decode(&s); err != nil {
		return err
	}
	return nil
}

// Delete alters the record of the file from DB
func (s *File) Delete(ctx context.Context) (*mongo.DeleteResult, error) {
	res, err := GetDB().Collection(collNameStatic).DeleteOne(ctx, s.getFilter())
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ListFilenamesWhere does what it should
func ListFilenamesWhere(ctx context.Context, dir, ext string) ([]string, error) {
	filter := bson.M{"dir": dir}
	if ext != "*" {
		filter["ext"] = ext
	}
	opts := options.Find().SetProjection(bson.D{{"name", 1}, {"ext", 1}})

	cur, err := GetDB().Collection(collNameStatic).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	filenames := make([]string, 0)
	for cur.Next(ctx) {
		var file File
		if err := cur.Decode(&file); err != nil {
			return nil, err
		}
		filenames = append(filenames, file.Name+"."+file.Ext)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return filenames, nil
}
