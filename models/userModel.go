package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	ID						primitive.ObjectID `bson:"_id"`
	First_Name		*string						`json:"first_name" validate:"required, min=2, max=100"`
	Last_Name			*string						`json:"last_name" validate:"required, min=2, max=100"`
	Password			*string						`json:"password" validate:"required, min=6"`
	Email					*string						`json:"email" validate:"required, email"`
	Avatar				*string						`json:"avatar"`
	Phone					*string						`json:"phone" validate:"required, min=11, max=11"`
	Token					*string						`json:"token"`
	Refresh_Token *string 					`json:"refresh_token"`
	Created_at		time.Time					`json:"created_at"`
	Updated_at		time.Time					`json:"updated_at"`
	User_id				string						`json:"user_id"`
}