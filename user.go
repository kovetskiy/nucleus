package main

type user struct {
	Name       string                 `bson:"username"`
	Info       map[string]interface{} `bson:"userinfo"`
	Token      string                 `bson:"token"`
	TokenDate  int64                  `bson:"token_date"`
	CreateDate int64                  `bson:"create_date"`
}
