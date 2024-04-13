// Package metadata contains the metadata of this plugin
package metadata

import "github.com/cloudnative-pg/cnpg-i/pkg/identity"

// Data is the metadata of this plugin
// LEO: does this really belong here?
var Data = identity.GetPluginMetadataResponse{
	Name:          "objstore-backup.dougkirkley",
	Version:       "0.0.1",
	DisplayName:   "CNPG-I plugin to backup and recover using an Object Store",
	ProjectUrl:    "https://github.com/dougkirkley/plugin-objstore-backup",
	RepositoryUrl: "https://github.com/dougkirkley/plugin-objstore-backup",
	License:       "Apache 2",
	LicenseUrl:    "https://github.com/dougkirkley/plugin-objstore-backup/blob/main/LICENSE",
	Maturity:      "alpha",
	Vendor:        "Douglass Kirkley",
}
