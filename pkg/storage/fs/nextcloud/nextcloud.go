// Copyright 2018-2021 CERN
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

package nextcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	provider "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"
	types "github.com/cs3org/go-cs3apis/cs3/types/v1beta1"
	"github.com/cs3org/reva/pkg/storage"
	"github.com/cs3org/reva/pkg/storage/fs/registry"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

func init() {
	registry.Register("nextcloud", New)
}

type Config struct {
	EndPoint string `mapstructure:"end_point"` // e.g. "http://nextcloud/app/sciencemesh/do.php"
}
type nextcloud struct {
	endPoint string
}

func parseConfig(m map[string]interface{}) (*Config, error) {
	c := &Config{}
	if err := mapstructure.Decode(m, c); err != nil {
		err = errors.Wrap(err, "error decoding conf")
		return nil, err
	}
	return c, nil
}

// New returns an implementation to of the storage.FS interface that talks to
// a Nextcloud instance over http.
func New(m map[string]interface{}) (storage.FS, error) {
	conf, err := parseConfig(m)
	if err != nil {
		return nil, err
	}

	return NewNextcloud(conf)
}

func NewNextcloud(c *Config) (storage.FS, error) {
	return &nextcloud{
		endPoint: c.EndPoint, // e.g. "http://nextcloud/app/sciencemesh/do.php"
	}, nil
}

type Action struct {
	verb string
	argS string
}

func (nc *nextcloud) doUpload(r io.ReadCloser) error {
	// See https://github.com/pondersource/sciencemesh-nextcloud/issues/13
	endPoint := "http://nc/apps/sciencemesh/test"

	fmt.Printf("\nUPLOADING IT TO %s!\n\n", endPoint)

	// initialize http client
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, "http://api.example.com/v1/user", bytes.NewBuffer(json))
	if err != nil {
		panic(err)
	}

	// set the request header Content-Type for json
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, err2 := io.ReadAll(resp.Body)
	fmt.Printf("\nRESPONSE BODY %s %i!\n\n", body, resp.StatusCode)
	return err2
}

func (nc *nextcloud) do(a Action) (string, error) {
	fmt.Printf("\naction %s %s %s\n\n", nc.endPoint, a.verb, a.argS)
	return "printed", nil
}

func (nc *nextcloud) GetHome(ctx context.Context) (string, error) {
	return nc.do(Action{"GetHome", ""})
}
func (nc *nextcloud) CreateHome(ctx context.Context) error {
	_, err := nc.do(Action{"CreateHome", ""})
	return err
}
func (nc *nextcloud) CreateDir(ctx context.Context, fn string) error {
	_, err := nc.do(Action{"CreateDir", fn})
	return err
}
func (nc *nextcloud) Delete(ctx context.Context, ref *provider.Reference) error {
	s, _ := json.Marshal(ref)
	_, err := nc.do(Action{"Delete", string(s)})
	return err
}
func (nc *nextcloud) Move(ctx context.Context, oldRef, newRef *provider.Reference) error {
	s, _ := json.Marshal(newRef)
	_, err := nc.do(Action{"Move", string(s)})
	return err
}
func (nc *nextcloud) GetMD(ctx context.Context, ref *provider.Reference, mdKeys []string) (*provider.ResourceInfo, error) {
	s, _ := json.Marshal(ref)
	nc.do(Action{"GetMD", string(s)})

	fp := "some-file.txt"
	// example:
	// {
	// 	"type":1,
	// 	"id":{
	// 		"opaque_id":"fileid-einstein%2Ffile.txt"
	// 	},
	// 	"etag":"\"e13c8b47adc153eb32036b634072d4a8\"",
	// 	"mime_type":"text/plain; charset=utf-8",
	// 	"mtime":{
	// 		"seconds":1624369389
	// 	},
	// 	"path":"/file.txt",
	// 	"permission_set":{
	// 		"add_grant":true,
	// 		"create_container":true,
	// 		"delete":true,
	// 		"get_path":true,
	// 		"get_quota":true,
	// 		"initiate_file_download":true,
	// 		"initiate_file_upload":true,
	// 		"list_grants":true,
	// 		"list_container":true,
	// 		"list_file_versions":true,
	// 		"list_recycle":true,
	// 		"move":true,
	// 		"remove_grant":true,
	// 		"purge_recycle":true,
	// 		"restore_file_version":true,
	// 		"restore_recycle_item":true,
	// 		"stat":true,
	// 		"update_grant":true
	// 	},
	// 	"size":6990,
	// 	"owner":{
	// 		"idp":"cernbox.cern.ch",
	// 		"opaque_id":"4c510ada-c86b-4815-8820-42cdf82c3d51"
	// 	},
	// 	"arbitrary_metadata":{
	// 	}
	// }
	md := &provider.ResourceInfo{
		Id:            &provider.ResourceId{OpaqueId: "fileid-" + url.QueryEscape(fp)},
		Path:          fp,
		Type:          provider.ResourceType_RESOURCE_TYPE_FILE,
		Etag:          "some-etag",
		MimeType:      "application/octet-stream",
		Size:          0,
		PermissionSet: &provider.ResourcePermissions{
			// no permissions
		},
		Mtime: &types.Timestamp{
			Seconds: 1234567890,
		},
		Owner:             nil,
		ArbitraryMetadata: nil,
	}

	return md, nil
}
func (nc *nextcloud) ListFolder(ctx context.Context, ref *provider.Reference, mdKeys []string) ([]*provider.ResourceInfo, error) {
	s, _ := json.Marshal(ref)
	_, err := nc.do(Action{"ListFolder", string(s)})
	return nil, err
}

// Copied from https://github.com/cs3org/reva/blob/a8c61401b662d8e09175416c0556da8ef3ba8ed6/pkg/storage/utils/eosfs/upload.go#L77-L81
func (nc *nextcloud) InitiateUpload(ctx context.Context, ref *provider.Reference, uploadLength int64, metadata map[string]string) (map[string]string, error) {
	s, _ := json.Marshal(ref)
	_, err := nc.do(Action{"InitiateUpload", string(s)})
	return map[string]string{
		"simple": ref.GetPath(),
	}, err
}
func (nc *nextcloud) Upload(ctx context.Context, ref *provider.Reference, r io.ReadCloser) error {
	s, _ := json.Marshal(ref)
	nc.doUpload(r)
	_, err := nc.do(Action{"Upload", string(s)})
	return err
}
func (nc *nextcloud) Download(ctx context.Context, ref *provider.Reference) (io.ReadCloser, error) {
	s, _ := json.Marshal(ref)
	_, err := nc.do(Action{"Download", string(s)})
	return nil, err
}
func (nc *nextcloud) ListRevisions(ctx context.Context, ref *provider.Reference) ([]*provider.FileVersion, error) {
	s, _ := json.Marshal(ref)
	_, err := nc.do(Action{"ListRevisions", string(s)})
	return nil, err
}
func (nc *nextcloud) DownloadRevision(ctx context.Context, ref *provider.Reference, key string) (io.ReadCloser, error) {
	s, _ := json.Marshal(ref)
	_, err := nc.do(Action{"DownloadRevision", string(s)})
	return nil, err
}
func (nc *nextcloud) RestoreRevision(ctx context.Context, ref *provider.Reference, key string) error {
	s, _ := json.Marshal(ref)
	_, err := nc.do(Action{"RestoreRevision", string(s)})
	return err
}
func (nc *nextcloud) ListRecycle(ctx context.Context) ([]*provider.RecycleItem, error) {
	_, err := nc.do(Action{"ListRecycle", ""})
	return nil, err
}

// func (nc *nextcloud) RestoreRecycleItem(ctx context.Context, key string, restoreRef *provider.Reference) error {
func (nc *nextcloud) RestoreRecycleItem(ctx context.Context, key string, restoreRef string) error {
	s, _ := json.Marshal(restoreRef)
	_, err := nc.do(Action{"RestoreRecycleItem", string(s)})
	return err
}
func (nc *nextcloud) PurgeRecycleItem(ctx context.Context, key string) error {
	s, _ := json.Marshal(key)
	_, err := nc.do(Action{"PurgeRecycleItem", string(s)})
	return err
}
func (nc *nextcloud) EmptyRecycle(ctx context.Context) error {
	_, err := nc.do(Action{"EmptyRecycle", ""})
	return err
}
func (nc *nextcloud) GetPathByID(ctx context.Context, id *provider.ResourceId) (string, error) {
	s, _ := json.Marshal(id)
	_, err := nc.do(Action{"GetPathByID", string(s)})
	return "sorry", err
}

func (nc *nextcloud) AddGrant(ctx context.Context, ref *provider.Reference, g *provider.Grant) error {
	s, _ := json.Marshal(ref)
	_, err := nc.do(Action{"AddGrant", string(s)})
	return err
}
func (nc *nextcloud) RemoveGrant(ctx context.Context, ref *provider.Reference, g *provider.Grant) error {
	s, _ := json.Marshal(ref)
	_, err := nc.do(Action{"RemoveGrant", string(s)})
	return err
}
func (nc *nextcloud) UpdateGrant(ctx context.Context, ref *provider.Reference, g *provider.Grant) error {
	s, _ := json.Marshal(ref)
	_, err := nc.do(Action{"UpdateGrant", string(s)})
	return err
}
func (nc *nextcloud) ListGrants(ctx context.Context, ref *provider.Reference) ([]*provider.Grant, error) {
	s, _ := json.Marshal(ref)
	_, err := nc.do(Action{"ListGrants", string(s)})
	return nil, err
}
func (nc *nextcloud) GetQuota(ctx context.Context) (uint64, uint64, error) {
	_, err := nc.do(Action{"GetQuota", ""})
	return 0, 0, err
}
func (nc *nextcloud) CreateReference(ctx context.Context, path string, targetURI *url.URL) error {
	_, err := nc.do(Action{"CreateReference", path})
	return err
}
func (nc *nextcloud) Shutdown(ctx context.Context) error {
	_, err := nc.do(Action{"Shutdown", ""})
	return err
}
func (nc *nextcloud) SetArbitraryMetadata(ctx context.Context, ref *provider.Reference, md *provider.ArbitraryMetadata) error {
	s, _ := json.Marshal(md)
	_, err := nc.do(Action{"SetArbitraryMetadata", string(s)})
	return err
}
func (nc *nextcloud) UnsetArbitraryMetadata(ctx context.Context, ref *provider.Reference, keys []string) error {
	s, _ := json.Marshal(ref)
	_, err := nc.do(Action{"UnsetArbitraryMetadata", string(s)})
	return err
}
