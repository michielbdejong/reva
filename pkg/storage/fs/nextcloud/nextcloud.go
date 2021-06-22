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
	"github.com/cs3org/reva/pkg/storage"
	"github.com/cs3org/reva/pkg/storage/fs/registry"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	codes "google.golang.org/grpc/codes"
	gstatus "google.golang.org/grpc/status"
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

func (nc *nextcloud) do(a Action) (string, error) {
	b, err := json.Marshal(a)
	fmt.Println("action %s\n", b)
	resp, err := http.Post(nc.endPoint, "application/json", bytes.NewReader(b))
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return string(body), err
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
	return gstatus.Errorf(codes.Unimplemented, "method not implemented")
}
func (nc *nextcloud) Move(ctx context.Context, oldRef, newRef *provider.Reference) error {
	return gstatus.Errorf(codes.Unimplemented, "method not implemented")
}
func (nc *nextcloud) GetMD(ctx context.Context, ref *provider.Reference, mdKeys []string) (*provider.ResourceInfo, error) {
	return &provider.ResourceInfo{}, nil
}
func (nc *nextcloud) ListFolder(ctx context.Context, ref *provider.Reference, mdKeys []string) ([]*provider.ResourceInfo, error) {
	return nil, gstatus.Errorf(codes.Unimplemented, "method not implemented")
}

// Copied from https://github.com/cs3org/reva/blob/a8c61401b662d8e09175416c0556da8ef3ba8ed6/pkg/storage/utils/eosfs/upload.go#L77-L81
func (fs *nextcloud) InitiateUpload(ctx context.Context, ref *provider.Reference, uploadLength int64, metadata map[string]string) (map[string]string, error) {
	return map[string]string{
		"simple": ref.GetPath(),
	}, nil
}
func (nc *nextcloud) Upload(ctx context.Context, ref *provider.Reference, r io.ReadCloser) error {
	fmt.Println("upload! %s", r)
	return nil
}
func (nc *nextcloud) Download(ctx context.Context, ref *provider.Reference) (io.ReadCloser, error) {
	return nil, gstatus.Errorf(codes.Unimplemented, "method not implemented")
}
func (nc *nextcloud) ListRevisions(ctx context.Context, ref *provider.Reference) ([]*provider.FileVersion, error) {
	return nil, gstatus.Errorf(codes.Unimplemented, "method not implemented")
}
func (nc *nextcloud) DownloadRevision(ctx context.Context, ref *provider.Reference, key string) (io.ReadCloser, error) {
	return nil, gstatus.Errorf(codes.Unimplemented, "method not implemented")
}
func (nc *nextcloud) RestoreRevision(ctx context.Context, ref *provider.Reference, key string) error {
	return gstatus.Errorf(codes.Unimplemented, "method not implemented")
}
func (nc *nextcloud) ListRecycle(ctx context.Context) ([]*provider.RecycleItem, error) {
	return nil, gstatus.Errorf(codes.Unimplemented, "method not implemented")
}
func (nc *nextcloud) RestoreRecycleItem(ctx context.Context, key, restorePath string) error {
	return gstatus.Errorf(codes.Unimplemented, "method not implemented")
}
func (nc *nextcloud) PurgeRecycleItem(ctx context.Context, key string) error {
	return gstatus.Errorf(codes.Unimplemented, "method not implemented")
}
func (nc *nextcloud) EmptyRecycle(ctx context.Context) error {
	return gstatus.Errorf(codes.Unimplemented, "method not implemented")
}
func (nc *nextcloud) GetPathByID(ctx context.Context, id *provider.ResourceId) (string, error) {
	return "sorry", gstatus.Errorf(codes.Unimplemented, "method not implemented")
}
func (nc *nextcloud) AddGrant(ctx context.Context, ref *provider.Reference, g *provider.Grant) error {
	return gstatus.Errorf(codes.Unimplemented, "method not implemented")
}
func (nc *nextcloud) RemoveGrant(ctx context.Context, ref *provider.Reference, g *provider.Grant) error {
	return gstatus.Errorf(codes.Unimplemented, "method not implemented")
}
func (nc *nextcloud) UpdateGrant(ctx context.Context, ref *provider.Reference, g *provider.Grant) error {
	return gstatus.Errorf(codes.Unimplemented, "method not implemented")
}
func (nc *nextcloud) ListGrants(ctx context.Context, ref *provider.Reference) ([]*provider.Grant, error) {
	return nil, gstatus.Errorf(codes.Unimplemented, "method not implemented")
}
func (nc *nextcloud) GetQuota(ctx context.Context) (uint64, uint64, error) {
	return 0, 0, gstatus.Errorf(codes.Unimplemented, "method not implemented")
}
func (nc *nextcloud) CreateReference(ctx context.Context, path string, targetURI *url.URL) error {
	return gstatus.Errorf(codes.Unimplemented, "method not implemented")
}
func (nc *nextcloud) Shutdown(ctx context.Context) error {
	return gstatus.Errorf(codes.Unimplemented, "method not implemented")
}
func (nc *nextcloud) SetArbitraryMetadata(ctx context.Context, ref *provider.Reference, md *provider.ArbitraryMetadata) error {
	return gstatus.Errorf(codes.Unimplemented, "method not implemented")
}
func (nc *nextcloud) UnsetArbitraryMetadata(ctx context.Context, ref *provider.Reference, keys []string) error {
	return gstatus.Errorf(codes.Unimplemented, "method not implemented")
}
