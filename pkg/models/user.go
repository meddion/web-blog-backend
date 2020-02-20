package models

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

const collNameUser = "users"

type User struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name         string             `json:"name" bson:"name,omitempty"`
	Password     string             `json:"password" bson:"password,omitempty"`
	CreationTime int64              `json:"creation_time" bson:"creation_time,omitempty"`
}

func (u *User) Update(ctx context.Context) (err error) {
	if u.Password != "" {
		if u.Password, err = hashPassword(u.Password); err != nil {
			return
		}
	}
	u.CreationTime = 0
	_, err = db.Collection(collNameUser).UpdateOne(
		ctx,
		bson.M{"_id": u.ID},
		bson.M{"$set": u},
	)
	return
}

func (u *User) Create(ctx context.Context) error {
	var err error
	if u.Password, err = hashPassword(u.Password); err != nil {
		return err
	}
	u.CreationTime = time.Now().Unix()
	result, err := db.Collection(collNameUser).InsertOne(ctx, u)
	if err != nil {
		return err
	}
	if objectID, ok := result.InsertedID.(primitive.ObjectID); ok {
		u.ID = objectID
		return nil
	}
	return errors.New("on retrieving undefined id type")
}

func DeleteUserByID(ctx context.Context, id primitive.ObjectID) (err error) {
	_, err = db.Collection(collNameUser).DeleteOne(ctx, bson.D{{"_id", id}})
	return
}

func (u *User) Get(ctx context.Context) error {
	//ctx, _ = context.WithTimeout(ctx, 3*time.Second)
	return db.Collection(collNameUser).FindOne(ctx, bson.D{{"name", u.Name}}).Decode(u)
}

func GetUserInfoByName(ctx context.Context, name string) (*User, error) {
	user := &User{}
	opts := options.FindOne().SetProjection(bson.D{
		{"_id", 0},
		{"password", 0},
		{"creation_time", 0},
	}) // mark by 0 to exclude a field
	err := db.Collection(collNameUser).FindOne(ctx, bson.D{{"name", name}}, opts).Decode(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func CompareHashAndPassword(hash, password string) (bool, error) {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// Validators
func (u *User) ValidateLoginForm() error {
	if err := u.ValidateName(); err != nil {
		return err
	}
	return u.ValidatePassword()
}

func (u *User) ValidateSignupForm(ctx context.Context) error {
	if err := u.ValidateName(); err != nil {
		return err
	}
	if err := u.ValidatePassword(); err != nil {
		return err
	}
	if !u.IsNameUnique(ctx) {
		return errors.New("on receiving not a unique username")
	}
	return nil
}

func (u *User) ValidateUpdateForm(ctx context.Context) error {
	switch {
	case u.Password != "":
		if err := u.ValidatePassword(); err != nil {
			return err
		}
	case u.Name != "":
		if err := u.ValidateName(); err != nil {
			return err
		}
		if !u.IsNameUnique(ctx) {
			return errors.New("on receiving not a unique username")
		}
	}
	return nil
}

func (u *User) ValidateName() error {
	if len(u.Name) < 3 || len(u.Name) > 32 {
		return errors.New("on receiving a username that is less than 3 or more than 32 chars long")
	}
	return nil
}

func (u *User) ValidatePassword() error {
	if len(u.Password) < 6 || len(u.Password) > 32 {
		return errors.New("on receiving a password that is less than 6 or more than 32 chars long")
	}
	return nil
}

func (u *User) IsNameUnique(ctx context.Context) bool {
	tempUser := &User{Name: u.Name}
	if err := tempUser.Get(ctx); err == mongo.ErrNoDocuments {
		return true
	}
	return false
}

// Helpers

// Assign sets all the non-empty fields (excepty ID, CreationTime) of newUser to u struct
func (u *User) Assign(newUser *User) {
	if newUser.Name != "" {
		u.Name = newUser.Name
	}
	if newUser.Password != "" {
		u.Password = newUser.Password
	}
}
