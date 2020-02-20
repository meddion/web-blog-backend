package models

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const collNamePost = "posts"

// Post struct is a model for a post collection
type Post struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title        string             `json:"title" bson:"title"`
	Content      string             `json:"content" bson:"content"`
	AuthorID     primitive.ObjectID `json:"author_id" bson:"author_id,omitempty"`
	CreationTime int64              `json:"creation_time" bson:"creation_time,omitempty"`
	LastEdited   int64              `json:"last_edited" bson:"last_edited,omitempty"`
}

type PostWithAuthor struct {
	Post   `bson:"inline"`
	Author User `json:"author" bson:"author,omitempty"`
}

func (p *Post) Update(ctx context.Context) (err error) {
	p.LastEdited = time.Now().Unix()
	_, err = GetDB().Collection(collNamePost).UpdateOne(ctx, bson.M{"_id": p.ID}, bson.M{"$set": p})
	return
}

func (p *Post) Save(ctx context.Context) error {
	p.CreationTime = time.Now().Unix()
	result, err := GetDB().Collection(collNamePost).InsertOne(ctx, p)
	if err != nil {
		return err
	}
	if objectID, ok := result.InsertedID.(primitive.ObjectID); ok {
		p.ID = objectID
		return nil
	}
	return errors.New("on retrieving undefined id type")
}

func DeletePostById(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	res, err := GetDB().Collection(collNamePost).DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}
	if res.DeletedCount < 1 {
		return errors.New("the resource was not found, thus nothing was deleted")
	}
	return nil
}

func GetPostByID(ctx context.Context, id string) (*Post, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	post := &Post{}
	err = GetDB().Collection(collNamePost).FindOne(ctx, bson.M{"_id": objectID}).Decode(post)
	return post, err
}

func GetTotalNumOfPosts(ctx context.Context, filter interface{}) (int64, error) {
	if filter == nil {
		filter = bson.D{{}}
	}
	totalNumOfPosts, err := GetDB().Collection(collNamePost).CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return totalNumOfPosts, nil
}

func CreatePostsPipeline(r *http.Request, pageSize, pageNum int64) interface{} {
	sort := -1 // asc
	if r.URL.Query().Get("sortByDate") == "desc" {
		sort = 1 // desc
	}
	return bson.A{
		bson.M{"$sort": bson.M{"creation_time": sort}},
		bson.M{"$skip": (pageNum - 1) * pageSize},
		bson.M{"$limit": pageSize},
		/* bson.M{"$lookup": bson.M{
			"localField":   "author_id",
			"from":         "users",
			"foreignField": "_id",
			"as":           "author",
		}},
		bson.M{"$unwind": "$author"},
		bson.M{"$project": bson.M{
			"author_id":            0,
			"author.password":      0,
			"author.creation_time": 0,
		}}, */
	}
}

func GetPosts(ctx context.Context, pipeline interface{}) ([]*PostWithAuthor, error) {
	var posts []*PostWithAuthor

	cur, err := GetDB().Collection(collNamePost).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var elem PostWithAuthor
		err := cur.Decode(&elem)
		if err != nil {
			return nil, err
		}
		posts = append(posts, &elem)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}
