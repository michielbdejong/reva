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

package nextcloud_test

import (
	"context"
	"encoding/json"
	// "net/http"
	"os"

	"google.golang.org/grpc/metadata"

	userpb "github.com/cs3org/go-cs3apis/cs3/identity/user/v1beta1"
	provider "github.com/cs3org/go-cs3apis/cs3/storage/provider/v1beta1"
	"github.com/cs3org/reva/pkg/auth/scope"
	ctxpkg "github.com/cs3org/reva/pkg/ctx"
	"github.com/cs3org/reva/pkg/storage/fs/nextcloud"
	jwt "github.com/cs3org/reva/pkg/token/manager/jwt"
	"github.com/cs3org/reva/tests/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Nextcloud", func() {
	var (
		ctx     context.Context
		options map[string]interface{}
		tmpRoot string
		user    = &userpb.User{
			Id: &userpb.UserId{
				Idp:      "0.0.0.0:19000",
				OpaqueId: "f7fbf8c8-139b-4376-b307-cf0a8c2d0d9c",
				Type:     userpb.UserType_USER_TYPE_PRIMARY,
			},
			Username: "tester",
		}
	)

	BeforeEach(func() {
		var err error
		tmpRoot, err := helpers.TempDir("reva-unit-tests-*-root")
		Expect(err).ToNot(HaveOccurred())

		options = map[string]interface{}{
			"root":         tmpRoot,
			"enable_home":  true,
			"share_folder": "/Shares",
		}

		ctx = context.Background()

		// Add auth token
		tokenManager, err := jwt.New(map[string]interface{}{"secret": "changemeplease"})
		Expect(err).ToNot(HaveOccurred())
		scope, err := scope.AddOwnerScope(nil)
		Expect(err).ToNot(HaveOccurred())
		t, err := tokenManager.MintToken(ctx, user, scope)
		Expect(err).ToNot(HaveOccurred())
		ctx = ctxpkg.ContextSetToken(ctx, t)
		ctx = metadata.AppendToOutgoingContext(ctx, ctxpkg.TokenHeader, t)
		ctx = ctxpkg.ContextSetUser(ctx, user)
	})

	AfterEach(func() {
		if tmpRoot != "" {
			os.RemoveAll(tmpRoot)
		}
	})

	Describe("New", func() {
		It("returns a new instance", func() {
			_, err := nextcloud.New(options)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	// 	GetHome(ctx context.Context) (string, error)
	Describe("GetHome", func() {
		It("calls the GetHome endpoint", func() {
			nc, _ := nextcloud.NewStorageDriver(&nextcloud.StorageDriverConfig{
				EndPoint: "http://mock.com/apps/sciencemesh/",
				MockHTTP: true,
			})
			called := make([]string, 0)

			h := nextcloud.GetNextcloudServerMock(&called)
			mock, teardown := nextcloud.TestingHTTPClient(h)
			defer teardown()
			nc.SetHTTPClient(mock)
			home, err := nc.GetHome(ctx)
			Expect(home).To(Equal("yes we are"))
			Expect(err).ToNot(HaveOccurred())
			Expect(len(called)).To(Equal(1))
			Expect(called[0]).To(Equal("POST /apps/sciencemesh/~tester/api/GetHome "))
		})
	})

	// CreateHome(ctx context.Context) error
	Describe("CreateHome", func() {
		It("calls the CreateHome endpoint", func() {
			nc, _ := nextcloud.NewStorageDriver(&nextcloud.StorageDriverConfig{
				EndPoint: "http://mock.com/apps/sciencemesh/",
				MockHTTP: true,
			})
			called := make([]string, 0)
			h := nextcloud.GetNextcloudServerMock(&called)
			mock, teardown := nextcloud.TestingHTTPClient(h)
			defer teardown()
			nc.SetHTTPClient(mock)
			err := nc.CreateHome(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(called)).To(Equal(1))
			Expect(called[0]).To(Equal("POST /apps/sciencemesh/~tester/api/CreateHome "))
		})
	})

	// CreateDir(ctx context.Context, ref *provider.Reference) error
	Describe("CreateDir", func() {
		It("calls the CreateDir endpoint", func() {
			nc, _ := nextcloud.NewStorageDriver(&nextcloud.StorageDriverConfig{
				EndPoint: "http://mock.com/apps/sciencemesh/",
				MockHTTP: true,
			})
			called := make([]string, 0)
			h := nextcloud.GetNextcloudServerMock(&called)
			mock, teardown := nextcloud.TestingHTTPClient(h)
			defer teardown()
			nc.SetHTTPClient(mock)
			// https://github.com/cs3org/go-cs3apis/blob/970eec3/cs3/storage/provider/v1beta1/resources.pb.go#L550-L561
			ref := &provider.Reference{
				ResourceId: &provider.ResourceId{
          StorageId: "storage-id",
					OpaqueId: "opaque-id",
				},
				Path: "/some/path",
			}
			err := nc.CreateDir(ctx, ref)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(called)).To(Equal(1))
			Expect(called[0]).To(Equal("POST /apps/sciencemesh/~tester/api/CreateDir {\"resource_id\":{\"storage_id\":\"storage-id\",\"opaque_id\":\"opaque-id\"},\"path\":\"/some/path\"}"))
		})
	})

	// Delete(ctx context.Context, ref *provider.Reference) error
	Describe("Delete", func() {
		It("calls the Delete endpoint", func() {
			nc, _ := nextcloud.NewStorageDriver(&nextcloud.StorageDriverConfig{
				EndPoint: "http://mock.com/apps/sciencemesh/",
				MockHTTP: true,
			})
			called := make([]string, 0)
			h := nextcloud.GetNextcloudServerMock(&called)
			mock, teardown := nextcloud.TestingHTTPClient(h)
			defer teardown()
			nc.SetHTTPClient(mock)
			// https://github.com/cs3org/go-cs3apis/blob/970eec3/cs3/storage/provider/v1beta1/resources.pb.go#L550-L561
			ref := &provider.Reference{
				ResourceId: &provider.ResourceId{
          StorageId: "storage-id",
					OpaqueId: "opaque-id",
				},
				Path: "/some/path",
			}
			err := nc.Delete(ctx, ref)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(called)).To(Equal(1))
			Expect(called[0]).To(Equal("POST /apps/sciencemesh/~tester/api/Delete {\"resource_id\":{\"storage_id\":\"storage-id\",\"opaque_id\":\"opaque-id\"},\"path\":\"/some/path\"}"))
		})
	})

	// Move(ctx context.Context, oldRef, newRef *provider.Reference) error
	Describe("Move", func() {
		It("calls the Move endpoint", func() {
			nc, _ := nextcloud.NewStorageDriver(&nextcloud.StorageDriverConfig{
				EndPoint: "http://mock.com/apps/sciencemesh/",
				MockHTTP: true,
			})
			called := make([]string, 0)
			h := nextcloud.GetNextcloudServerMock(&called)
			mock, teardown := nextcloud.TestingHTTPClient(h)
			defer teardown()
			nc.SetHTTPClient(mock)
			// https://github.com/cs3org/go-cs3apis/blob/970eec3/cs3/storage/provider/v1beta1/resources.pb.go#L550-L561
			ref1 := &provider.Reference{
				ResourceId: &provider.ResourceId{
          StorageId: "storage-id-1",
					OpaqueId: "opaque-id-1",
				},
				Path: "/some/old/path",
			}
			ref2 := &provider.Reference{
				ResourceId: &provider.ResourceId{
          StorageId: "storage-id-2",
					OpaqueId: "opaque-id-2",
				},
				Path: "/some/new/path",
			}
			err := nc.Move(ctx, ref1, ref2)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(called)).To(Equal(1))
			Expect(called[0]).To(Equal("POST /apps/sciencemesh/~tester/api/Move {\"from\":{\"resource_id\":{\"storage_id\":\"storage-id-1\",\"opaque_id\":\"opaque-id-1\"},\"path\":\"/some/old/path\"},\"to\":{\"resource_id\":{\"storage_id\":\"storage-id-2\",\"opaque_id\":\"opaque-id-2\"},\"path\":\"/some/new/path\"}}"))
		})
	})

	// GetMD(ctx context.Context, ref *provider.Reference, mdKeys []string) (*provider.ResourceInfo, error)
	Describe("GetMD", func() {
		It("calls the GetMD endpoint", func() {
			nc, _ := nextcloud.NewStorageDriver(&nextcloud.StorageDriverConfig{
				EndPoint: "http://mock.com/apps/sciencemesh/",
				MockHTTP: true,
			})
			called := make([]string, 0)
			h := nextcloud.GetNextcloudServerMock(&called)
			mock, teardown := nextcloud.TestingHTTPClient(h)
			defer teardown()
			nc.SetHTTPClient(mock)
			// https://github.com/cs3org/go-cs3apis/blob/970eec3/cs3/storage/provider/v1beta1/resources.pb.go#L550-L561
			ref := &provider.Reference{
				ResourceId: &provider.ResourceId{
          StorageId: "storage-id",
					OpaqueId: "opaque-id",
				},
				Path: "/some/path",
			}
			mdKeys := []string{"val1", "val2", "val3"}
			result, err := nc.GetMD(ctx, ref, mdKeys)
			Expect(result.Etag).To(Equal("in-json-etag"))
			Expect(result.MimeType).To(Equal("in-json-mimetype"))
			resultJson, err := json.Marshal(result.ArbitraryMetadata)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(resultJson)).To(Equal("{\"metadata\":{\"foo\":\"bar\"}}"))
			Expect(err).ToNot(HaveOccurred())
			Expect(len(called)).To(Equal(1))
			Expect(called[0]).To(Equal("POST /apps/sciencemesh/~tester/api/GetMD {\"ref\":{\"resource_id\":{\"storage_id\":\"storage-id\",\"opaque_id\":\"opaque-id\"},\"path\":\"/some/path\"},\"mdKeys\":[\"val1\",\"val2\",\"val3\"]}"))
		})
	})

	// ListFolder(ctx context.Context, ref *provider.Reference, mdKeys []string) ([]*provider.ResourceInfo, error)
	Describe("ListFolder", func() {
		It("calls the ListFolder endpoint", func() {
			nc, _ := nextcloud.NewStorageDriver(&nextcloud.StorageDriverConfig{
				EndPoint: "http://mock.com/apps/sciencemesh/",
				MockHTTP: true,
			})
			called := make([]string, 0)
			h := nextcloud.GetNextcloudServerMock(&called)
			mock, teardown := nextcloud.TestingHTTPClient(h)
			defer teardown()
			nc.SetHTTPClient(mock)
			// https://github.com/cs3org/go-cs3apis/blob/970eec3/cs3/storage/provider/v1beta1/resources.pb.go#L550-L561
			ref := &provider.Reference{
				ResourceId: &provider.ResourceId{
          StorageId: "storage-id",
					OpaqueId: "opaque-id",
				},
				Path: "/some/path",
			}
			mdKeys := []string{"val1", "val2", "val3"}
			results, err := nc.ListFolder(ctx, ref, mdKeys)
			Expect(len(results)).To(Equal(1))
			Expect(results[0].Etag).To(Equal("in-json-etag"))
			Expect(results[0].MimeType).To(Equal("in-json-mimetype"))
			Expect(err).ToNot(HaveOccurred())
			Expect(len(called)).To(Equal(1))
			Expect(called[0]).To(Equal("POST /apps/sciencemesh/~tester/api/ListFolder {\"ref\":{\"resource_id\":{\"storage_id\":\"storage-id\",\"opaque_id\":\"opaque-id\"},\"path\":\"/some/path\"},\"mdKeys\":[\"val1\",\"val2\",\"val3\"]}"))
		})
	})

	// InitiateUpload(ctx context.Context, ref *provider.Reference, uploadLength int64, metadata map[string]string) (map[string]string, error)
	Describe("InitiateUpload", func() {
		It("calls the ListFolder endpoint", func() {
			nc, _ := nextcloud.NewStorageDriver(&nextcloud.StorageDriverConfig{
				EndPoint: "http://mock.com/apps/sciencemesh/",
				MockHTTP: true,
			})
			called := make([]string, 0)
			h := nextcloud.GetNextcloudServerMock(&called)
			mock, teardown := nextcloud.TestingHTTPClient(h)
			defer teardown()
			nc.SetHTTPClient(mock)
			// https://github.com/cs3org/go-cs3apis/blob/970eec3/cs3/storage/provider/v1beta1/resources.pb.go#L550-L561
			ref := &provider.Reference{
				ResourceId: &provider.ResourceId{
          StorageId: "storage-id",
					OpaqueId: "opaque-id",
				},
				Path: "/some/path",
			}
			uploadLength := int64(12345)
			metadata := map[string]string{
				"key1": "val1",
				"key2": "val2",
				"key3": "val3",
			}
			results, err := nc.InitiateUpload(ctx, ref, uploadLength, metadata)
			Expect(err).ToNot(HaveOccurred())
			Expect(results).To(Equal(map[string]string{
				"not": "sure",
				"what": "should be",
				"returned": "here",
			}))
			Expect(called[0]).To(Equal("POST /apps/sciencemesh/~tester/api/InitiateUpload {\"ref\":{\"resource_id\":{\"storage_id\":\"storage-id\",\"opaque_id\":\"opaque-id\"},\"path\":\"/some/path\"},\"uploadLength\":12345,\"metadata\":{\"key1\":\"val1\",\"key2\":\"val2\",\"key3\":\"val3\"}}"))
		})
	})

	// Upload(ctx context.Context, ref *provider.Reference, r io.ReadCloser) error
	// Download(ctx context.Context, ref *provider.Reference) (io.ReadCloser, error)
	// ListRevisions(ctx context.Context, ref *provider.Reference) ([]*provider.FileVersion, error)
	// DownloadRevision(ctx context.Context, ref *provider.Reference, key string) (io.ReadCloser, error)
	// RestoreRevision(ctx context.Context, ref *provider.Reference, key string) error
	// ListRecycle(ctx context.Context, key, path string) ([]*provider.RecycleItem, error)
	// RestoreRecycleItem(ctx context.Context, key, path string, restoreRef *provider.Reference) error
	// PurgeRecycleItem(ctx context.Context, key, path string) error
	// EmptyRecycle(ctx context.Context) error
	// GetPathByID(ctx context.Context, id *provider.ResourceId) (string, error)
	// AddGrant(ctx context.Context, ref *provider.Reference, g *provider.Grant) error
	// DenyGrant(ctx context.Context, ref *provider.Reference, g *provider.Grantee) error
	// RemoveGrant(ctx context.Context, ref *provider.Reference, g *provider.Grant) error
	// UpdateGrant(ctx context.Context, ref *provider.Reference, g *provider.Grant) error
	// ListGrants(ctx context.Context, ref *provider.Reference) ([]*provider.Grant, error)
	// GetQuota(ctx context.Context) (uint64, uint64, error)
	// CreateReference(ctx context.Context, path string, targetURI *url.URL) error
	// Shutdown(ctx context.Context) error
	// SetArbitraryMetadata(ctx context.Context, ref *provider.Reference, md *provider.ArbitraryMetadata) error
	// UnsetArbitraryMetadata(ctx context.Context, ref *provider.Reference, keys []string) error
	// ListStorageSpaces(ctx context.Context, filter []*provider.ListStorageSpacesRequest_Filter) ([]*provider.StorageSpace, error)

})
