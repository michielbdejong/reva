// Copyright 2018-2022 CERN
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// In applying this license, CERN does not waive the privileges and immunities
// granted to it by virtue of its status as an Intergovernmental Organization
// or submit itself to any jurisdiction.

package siteloc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/cs3org/reva/pkg/mentix/config"
	"github.com/cs3org/reva/pkg/mentix/meshdata"
	"github.com/rs/zerolog"
)

// HandleDefaultQuery processes a basic query.
func HandleDefaultQuery(meshData *meshdata.MeshData, params url.Values, _ *config.Configuration, _ *zerolog.Logger) (int, []byte, error) {
	// Convert the mesh data
	locData, err := convertMeshDataToLocationData(meshData)
	if err != nil {
		return http.StatusBadRequest, []byte{}, fmt.Errorf("unable to convert the mesh data to location data: %v", err)
	}

	// Marshal the location data as JSON
	data, err := json.MarshalIndent(locData, "", "\t")
	if err != nil {
		return http.StatusBadRequest, []byte{}, fmt.Errorf("unable to marshal the location data: %v", err)
	}

	return http.StatusOK, data, nil
}

func convertMeshDataToLocationData(meshData *meshdata.MeshData) ([]*SiteLocation, error) {
	// Gather the locations of all sites
	locations := make([]*SiteLocation, 0, len(meshData.Operators)*3)
	for _, op := range meshData.Operators {
		for _, site := range op.Sites {
			locations = append(locations, &SiteLocation{
				SiteID:    site.ID,
				FullName:  site.FullName,
				Longitude: site.Longitude,
				Latitude:  site.Latitude,
			})
		}
	}

	return locations, nil
}
