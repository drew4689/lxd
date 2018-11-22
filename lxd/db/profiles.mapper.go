// Note: the code below was backported from the 3.x master branch, and it's
//       mostly cut and paste from there to avoid re-inventing that logic.

package db

// The code below was generated by lxd-generate - DO NOT EDIT!

import (
	"github.com/lxc/lxd/lxd/db/cluster"
	"github.com/lxc/lxd/lxd/db/query"
	"github.com/lxc/lxd/shared/api"
	"github.com/pkg/errors"
)

var _ = api.ServerEnvironment{}

var profileObjects = cluster.RegisterStmt(`
SELECT profiles.id, profiles.name, coalesce(profiles.description, '')
  FROM profiles ORDER BY profiles.name
`)

var profileConfigRef = cluster.RegisterStmt(`
SELECT profiles.name,
       profiles_config.key,
       profiles_config.value
FROM profiles_config
JOIN profiles ON profiles.id=profiles_config.profile_id
`)

var profileDevicesRef = cluster.RegisterStmt(`
SELECT profiles.name,
       profiles_devices.name,
       profiles_devices.type,
       coalesce(profiles_devices_config.key, ''),
       coalesce(profiles_devices_config.value, '')
FROM profiles_devices
LEFT OUTER JOIN profiles_devices_config
   ON profiles_devices_config.profile_device_id=profiles_devices.id
 JOIN profiles ON profiles.id=profiles_devices.profile_id
`)

// ProfileList returns all available profiles.
func (c *ClusterTx) ProfileList() ([]Profile, error) {
	// Result slice.
	objects := make([]Profile, 0)

	stmt := c.stmt(profileObjects)
	args := []interface{}{}

	// Dest function for scanning a row.
	dest := func(i int) []interface{} {
		objects = append(objects, Profile{})
		return []interface{}{
			&objects[i].ID,
			&objects[i].Name,
			&objects[i].Description,
		}
	}

	// Select.
	err := query.SelectObjects(stmt, dest, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch profiles")
	}

	// Fill field Config.
	configObjects, err := c.ProfileConfigRef()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch field Config")
	}

	for i := range objects {
		value := configObjects[objects[i].Name]
		if value == nil {
			value = map[string]string{}
		}
		objects[i].Config = value
	}

	// Fill field Devices.
	devicesObjects, err := c.ProfileDevicesRef()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch field Devices")
	}

	for i := range objects {
		value := devicesObjects[objects[i].Name]
		if value == nil {
			value = map[string]map[string]string{}
		}
		objects[i].Devices = value
	}

	return objects, nil
}

// ProfileConfigRef returns entities used by profiles.
func (c *ClusterTx) ProfileConfigRef() (map[string]map[string]string, error) {
	// Result slice.
	objects := make([]struct {
		Name  string
		Key   string
		Value string
	}, 0)

	// Pick the prepared statement and arguments to use based on active criteria.
	stmt := c.stmt(profileConfigRef)
	args := []interface{}{}

	// Dest function for scanning a row.
	dest := func(i int) []interface{} {
		objects = append(objects, struct {
			Name  string
			Key   string
			Value string
		}{})
		return []interface{}{
			&objects[i].Name,
			&objects[i].Key,
			&objects[i].Value,
		}
	}

	// Select.
	err := query.SelectObjects(stmt, dest, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch  ref for profiles")
	}

	// Build index by primary name.
	index := map[string]map[string]string{}

	for _, object := range objects {
		item, ok := index[object.Name]
		if !ok {
			item = map[string]string{}
		}

		item[object.Key] = object.Value
	}

	return index, nil
}

// ProfileDevicesRef returns entities used by profiles.
func (c *ClusterTx) ProfileDevicesRef() (map[string]map[string]map[string]string, error) {
	// Result slice.
	objects := make([]struct {
		Name   string
		Device string
		Type   int
		Key    string
		Value  string
	}, 0)

	stmt := c.stmt(profileDevicesRef)
	args := []interface{}{}

	// Dest function for scanning a row.
	dest := func(i int) []interface{} {
		objects = append(objects, struct {
			Name   string
			Device string
			Type   int
			Key    string
			Value  string
		}{})
		return []interface{}{
			&objects[i].Name,
			&objects[i].Device,
			&objects[i].Type,
			&objects[i].Key,
			&objects[i].Value,
		}
	}

	// Select.
	err := query.SelectObjects(stmt, dest, args...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to fetch  ref for profiles")
	}

	// Build index by primary name.
	index := map[string]map[string]map[string]string{}

	for _, object := range objects {
		item, ok := index[object.Name]
		if !ok {
			item = map[string]map[string]string{}
		}

		index[object.Name] = item
		config, ok := item[object.Device]
		if !ok {
			// First time we see this device, let's int the config
			// and add the type.
			deviceType, err := dbDeviceTypeToString(object.Type)
			if err != nil {
				return nil, errors.Wrapf(
					err, "unexpected device type code '%d'", object.Type)
			}
			config = map[string]string{}
			config["type"] = deviceType
			item[object.Device] = config
		}
		if object.Key != "" {
			config[object.Key] = object.Value
		}
	}

	return index, nil
}