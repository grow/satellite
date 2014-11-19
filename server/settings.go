package server

import (
	"encoding/json"
	"time"

	"appengine"
	"appengine/datastore"
)

type AppSetting struct {
	JsonData  string
	UpdatedAt time.Time
}

const EntityAppSetting = "AppSetting"

type Settings map[string]string

func GetSettings(c appengine.Context, keys []string, values []Settings) error {
	datastoreKeys := make([]*datastore.Key, len(keys))
	for i, key := range keys {
		datastoreKeys[i] = datastore.NewKey(c, EntityAppSetting, key, 0 /* intID */, nil /* parent */)
	}
	entities := make([]AppSetting, len(keys))
	err := datastore.GetMulti(c, datastoreKeys, entities)
	if err != nil {
		return err
	}

	for i := range keys {
		err := json.Unmarshal([]byte(entities[i].JsonData), &values[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func SetSettings(c appengine.Context, key string, settings Settings) error {
	datastoreKey := datastore.NewKey(c, EntityAppSetting, key, 0 /* intID */, nil /* parent */)
	jsonData, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	entity := &AppSetting{
		JsonData:  string(jsonData),
		UpdatedAt: time.Now(),
	}
	_, err = datastore.Put(c, datastoreKey, entity)
	return err
}
