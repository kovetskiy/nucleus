package main

import mgo "gopkg.in/mgo.v2"

type database struct {
	dsn      string
	session  *mgo.Session
	database *mgo.Database
	tokens   *mgo.Collection
}

func (db *database) connect() error {
	var err error
	session, err := mgo.Dial(db.dsn)
	if err != nil {
		return err
	}

	db.session = session

	db.database = db.session.DB("")
	db.tokens = db.database.C("tokens")

	return nil
}

func (db *database) ping() error {
	return db.session.Ping()
}
