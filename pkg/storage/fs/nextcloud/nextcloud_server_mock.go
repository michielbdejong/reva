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
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
)

// Response contains data for the Nextcloud mock server to respond
// and to switch to a new server state
type Response struct {
	code           int
	body           string
	newServerState string
}

const serverStateError = "ERROR"
const serverStateEmpty = "EMPTY"
const serverStateHome = "HOME"
const serverStateSubdir = "SUBDIR"
const serverStateNewdir = "NEWDIR"
const serverStateSubdirNewdir = "SUBDIR-NEWDIR"
const serverStateFileRestored = "FILE-RESTORED"
const serverStateGrantAdded = "GRANT-ADDED"
const serverStateGrantUpdated = "GRANT-UPDATED"
const serverStateRecycle = "RECYCLE"
const serverStateReference = "REFERENCE"
const serverStateMetadata = "METADATA"

var serverState = serverStateEmpty

var responses = map[string]Response{
	`POST /apps/sciencemesh/~einstein/api/AddGrant {"path":"/subdir"}`: {200, ``, serverStateGrantAdded},

	`POST /apps/sciencemesh/~einstein/api/CreateDir {"path":"/subdir"} EMPTY`:  {200, ``, serverStateSubdir},
	`POST /apps/sciencemesh/~einstein/api/CreateDir {"path":"/subdir"} HOME`:   {200, ``, serverStateSubdir},
	`POST /apps/sciencemesh/~einstein/api/CreateDir {"path":"/subdir"} NEWDIR`: {200, ``, serverStateSubdirNewdir},

	`POST /apps/sciencemesh/~einstein/api/CreateDir {"path":"/newdir"} EMPTY`:  {200, ``, serverStateNewdir},
	`POST /apps/sciencemesh/~einstein/api/CreateDir {"path":"/newdir"} HOME`:   {200, ``, serverStateNewdir},
	`POST /apps/sciencemesh/~einstein/api/CreateDir {"path":"/newdir"} SUBDIR`: {200, ``, serverStateSubdirNewdir},

	`POST /apps/sciencemesh/~einstein/api/CreateHome `:   {200, ``, serverStateHome},
	`POST /apps/sciencemesh/~einstein/api/CreateHome {}`: {200, ``, serverStateHome},

	`POST /apps/sciencemesh/~einstein/api/CreateReference {"path":"/Shares/reference"}`: {200, `[]`, serverStateReference},

	`POST /apps/sciencemesh/~einstein/api/Delete {"path":"/subdir"}`: {200, ``, serverStateRecycle},

	`POST /apps/sciencemesh/~einstein/api/EmptyRecycle `: {200, ``, serverStateEmpty},

	`POST /apps/sciencemesh/~einstein/api/GetMD {"path":"/"} EMPTY`: {404, ``, serverStateEmpty},
	`POST /apps/sciencemesh/~einstein/api/GetMD {"path":"/"} HOME`:  {200, `{ "size": 1 }`, serverStateHome},

	`POST /apps/sciencemesh/~einstein/api/GetMD {"path":"/newdir"} EMPTY`:         {404, ``, serverStateEmpty},
	`POST /apps/sciencemesh/~einstein/api/GetMD {"path":"/newdir"} HOME`:          {404, ``, serverStateHome},
	`POST /apps/sciencemesh/~einstein/api/GetMD {"path":"/newdir"} SUBDIR`:        {404, ``, serverStateSubdir},
	`POST /apps/sciencemesh/~einstein/api/GetMD {"path":"/newdir"} NEWDIR`:        {200, `{ "size": 1 }`, serverStateNewdir},
	`POST /apps/sciencemesh/~einstein/api/GetMD {"path":"/newdir"} SUBDIR-NEWDIR`: {200, `{ "size": 1 }`, serverStateSubdirNewdir},

	`POST /apps/sciencemesh/~einstein/api/GetMD {"path":"/new_subdir"}`: {200, `{ "size": 1 }`, serverStateEmpty},

	`POST /apps/sciencemesh/~einstein/api/GetMD {"path":"/subdir"} EMPTY`:         {404, ``, serverStateEmpty},
	`POST /apps/sciencemesh/~einstein/api/GetMD {"path":"/subdir"} HOME`:          {404, ``, serverStateEmpty},
	`POST /apps/sciencemesh/~einstein/api/GetMD {"path":"/subdir"} NEWDIR`:        {404, ``, serverStateEmpty},
	`POST /apps/sciencemesh/~einstein/api/GetMD {"path":"/subdir"} RECYCLE`:       {404, ``, serverStateRecycle},
	`POST /apps/sciencemesh/~einstein/api/GetMD {"path":"/subdir"} SUBDIR`:        {200, `{ "size": 1 }`, serverStateEmpty},
	`POST /apps/sciencemesh/~einstein/api/GetMD {"path":"/subdir"} SUBDIR-NEWDIR`: {200, `{ "size": 1 }`, serverStateEmpty},
	`POST /apps/sciencemesh/~einstein/api/GetMD {"path":"/subdir"} METADATA`:      {200, `{ "size": 1, "metadata": { "foo": "bar" } }`, serverStateMetadata},

	`POST /apps/sciencemesh/~einstein/api/GetMD {"path":"/subdirRestored"} EMPTY`:         {404, ``, serverStateEmpty},
	`POST /apps/sciencemesh/~einstein/api/GetMD {"path":"/subdirRestored"} RECYCLE`:       {404, ``, serverStateRecycle},
	`POST /apps/sciencemesh/~einstein/api/GetMD {"path":"/subdirRestored"} SUBDIR`:        {404, ``, serverStateSubdir},
	`POST /apps/sciencemesh/~einstein/api/GetMD {"path":"/subdirRestored"} FILE-RESTORED`: {200, `{ "size": 1 }`, serverStateFileRestored},

	`POST /apps/sciencemesh/~einstein/api/GetMD {"path":"/versionedFile"} EMPTY`:         {200, `{ "size": 2 }`, serverStateEmpty},
	`POST /apps/sciencemesh/~einstein/api/GetMD {"path":"/versionedFile"} FILE-RESTORED`: {200, `{ "size": 1 }`, serverStateFileRestored},

	`POST /apps/sciencemesh/~einstein/api/GetPathByID {"storage_id":"00000000-0000-0000-0000-000000000000","opaque_id":"fileid-%2Fsubdir"}`: {200, "/subdir", serverStateEmpty},

	`POST /apps/sciencemesh/~einstein/api/InitiateUpload {"path":"/file"}`: {200, `{"simple": "yes","tus": "yes"}`, serverStateEmpty},

	`POST /apps/sciencemesh/~einstein/api/ListFolder {"path":"/"}`: {200, `["/subdir"]`, serverStateEmpty},

	`POST /apps/sciencemesh/~einstein/api/ListFolder {"path":"/Shares"} EMPTY`:     {404, ``, serverStateEmpty},
	`POST /apps/sciencemesh/~einstein/api/ListFolder {"path":"/Shares"} SUBDIR`:    {404, ``, serverStateSubdir},
	`POST /apps/sciencemesh/~einstein/api/ListFolder {"path":"/Shares"} REFERENCE`: {200, `["reference"]`, serverStateReference},

	`POST /apps/sciencemesh/~einstein/api/ListGrants {"path":"/subdir"} SUBDIR`:        {200, `[]`, serverStateEmpty},
	`POST /apps/sciencemesh/~einstein/api/ListGrants {"path":"/subdir"} GRANT-ADDED`:   {200, `[ { "stat": true, "move": true, "delete": false } ]`, serverStateEmpty},
	`POST /apps/sciencemesh/~einstein/api/ListGrants {"path":"/subdir"} GRANT-UPDATED`: {200, `[ { "stat": true, "move": true, "delete": true } ]`, serverStateEmpty},

	`POST /apps/sciencemesh/~einstein/api/ListRecycle  EMPTY`:   {200, `[]`, serverStateEmpty},
	`POST /apps/sciencemesh/~einstein/api/ListRecycle  RECYCLE`: {200, `["/subdir"]`, serverStateRecycle},

	`POST /apps/sciencemesh/~einstein/api/ListRevisions {"path":"/versionedFile"} EMPTY`:         {500, `[1]`, serverStateEmpty},
	`POST /apps/sciencemesh/~einstein/api/ListRevisions {"path":"/versionedFile"} FILE-RESTORED`: {500, `[1, 2]`, serverStateFileRestored},

	`POST /apps/sciencemesh/~einstein/api/Move {"from":"/subdir","to":"/new_subdir"}`: {200, ``, serverStateEmpty},

	`POST /apps/sciencemesh/~einstein/api/RemoveGrant {"path":"/subdir"} GRANT-ADDED`: {200, ``, serverStateGrantUpdated},

	`POST /apps/sciencemesh/~einstein/api/RestoreRecycleItem null`:                       {200, ``, serverStateSubdir},
	`POST /apps/sciencemesh/~einstein/api/RestoreRecycleItem {"path":"/subdirRestored"}`: {200, ``, serverStateFileRestored},

	`POST /apps/sciencemesh/~einstein/api/RestoreRevision {"path":"/versionedFile"}`: {200, ``, serverStateFileRestored},

	`POST /apps/sciencemesh/~einstein/api/SetArbitraryMetadata {"metadata":{"foo":"bar"}}`: {200, ``, serverStateMetadata},

	`POST /apps/sciencemesh/~einstein/api/UnsetArbitraryMetadata {"path":"/subdir"}`: {200, ``, serverStateSubdir},

	`POST /apps/sciencemesh/~einstein/api/UpdateGrant {"path":"/subdir"}`: {200, ``, serverStateGrantUpdated},

	`POST /apps/sciencemesh/~tester/api/GetHome `:    {200, `yes we are`, serverStateHome},
	`POST /apps/sciencemesh/~tester/api/CreateHome `: {201, ``, serverStateEmpty},
	`POST /apps/sciencemesh/~tester/api/CreateDir {"resource_id":{"storage_id":"storage-id","opaque_id":"opaque-id"},"path":"/some/path"}`:                                                                                                                        {201, ``, serverStateEmpty},
	`POST /apps/sciencemesh/~tester/api/Delete {"resource_id":{"storage_id":"storage-id","opaque_id":"opaque-id"},"path":"/some/path"}`:                                                                                                                           {200, ``, serverStateEmpty},
	`POST /apps/sciencemesh/~tester/api/Move {"oldRef":{"resource_id":{"storage_id":"storage-id-1","opaque_id":"opaque-id-1"},"path":"/some/old/path"},"newRef":{"resource_id":{"storage_id":"storage-id-2","opaque_id":"opaque-id-2"},"path":"/some/new/path"}}`: {200, ``, serverStateEmpty},
	`POST /apps/sciencemesh/~tester/api/GetMD {"ref":{"resource_id":{"storage_id":"storage-id","opaque_id":"opaque-id"},"path":"/some/path"},"mdKeys":["val1","val2","val3"]}`:                                                                                    {200, `{"opaque":{},"type":1,"id":{"opaque_id":"fileid-/some/path"},"checksum":{},"etag":"deadbeef","mime_type":"text/plain","mtime":{"seconds":1234567890},"path":"/some/path","permission_set":{},"size":12345,"canonical_metadata":{},"arbitrary_metadata":{"metadata":{"da":"ta","some":"arbi","trary":"meta"}}}`, serverStateEmpty},
	`POST /apps/sciencemesh/~tester/api/ListFolder {"ref":{"resource_id":{"storage_id":"storage-id","opaque_id":"opaque-id"},"path":"/some/path"},"mdKeys":["val1","val2","val3"]}`:                                                                               {200, `[{"opaque":{},"type":1,"id":{"opaque_id":"fileid-/some/path"},"checksum":{},"etag":"deadbeef","mime_type":"text/plain","mtime":{"seconds":1234567890},"path":"/some/path","permission_set":{},"size":12345,"canonical_metadata":{},"arbitrary_metadata":{"metadata":{"da":"ta","some":"arbi","trary":"meta"}}}]`, serverStateEmpty},
	`POST /apps/sciencemesh/~tester/api/InitiateUpload {"ref":{"resource_id":{"storage_id":"storage-id","opaque_id":"opaque-id"},"path":"/some/path"},"uploadLength":12345,"metadata":{"key1":"val1","key2":"val2","key3":"val3"}}`:                               {200, `{ "not":"sure", "what": "should be", "returned": "here" }`, serverStateEmpty},
	`PUT /apps/sciencemesh/~tester/api/Upload/some/file/path.txt shiny!`:                                                                                                                                                            {200, ``, serverStateEmpty},
	`GET /apps/sciencemesh/~tester/api/Download/some/file/path.txt `:                                                                                                                                                                {200, `the contents of the file`, serverStateEmpty},
	`POST /apps/sciencemesh/~tester/api/ListRevisions {"resource_id":{"storage_id":"storage-id","opaque_id":"opaque-id"},"path":"/some/path"}`:                                                                                      {200, `[{"opaque":{"map":{"some":{"value":"ZGF0YQ=="}}},"key":"version-12","size":12345,"mtime":1234567890,"etag":"deadb00f"},{"opaque":{"map":{"different":{"value":"c3R1ZmY="}}},"key":"asdf","size":12345,"mtime":1234567890,"etag":"deadbeef"}]`, serverStateEmpty},
	`GET /apps/sciencemesh/~tester/api/DownloadRevision/some%2Frevision/some/file/path.txt `:                                                                                                                                        {200, `the contents of that revision`, serverStateEmpty},
	`POST /apps/sciencemesh/~tester/api/RestoreRevision {"ref":{"resource_id":{"storage_id":"storage-id","opaque_id":"opaque-id"},"path":"some/file/path.txt"},"key":"asdf"}`:                                                       {200, ``, serverStateEmpty},
	`POST /apps/sciencemesh/~tester/api/ListRecycle {"key":"asdf","path":"/some/file.txt"}`:                                                                                                                                         {200, `[{"opaque":{},"key":"some-deleted-version","ref":{"resource_id":{},"path":"/some/file.txt"},"size":12345,"deletion_time":{"seconds":1234567890}}]`, serverStateEmpty},
	`POST /apps/sciencemesh/~tester/api/RestoreRecycleItem {"key":"asdf","path":"original/location/when/deleted.txt","restoreRef":{"resource_id":{"storage_id":"storage-id","opaque_id":"opaque-id"},"path":"some/file/path.txt"}}`: {200, ``, serverStateEmpty},
	`POST /apps/sciencemesh/~tester/api/PurgeRecycleItem {"key":"asdf","path":"original/location/when/deleted.txt"}`:                                                                                                                {200, ``, serverStateEmpty},
	`POST /apps/sciencemesh/~tester/api/EmptyRecycle `:                                                                                                                                                                              {200, ``, serverStateEmpty},
	`POST /apps/sciencemesh/~tester/api/GetPathByID {"storage_id":"storage-id","opaque_id":"opaque-id"}`:                                                                                                                            {200, `the/path/for/that/id.txt`, serverStateEmpty},
	`POST /apps/sciencemesh/~tester/api/AddGrant {"ref":{"resource_id":{"storage_id":"storage-id","opaque_id":"opaque-id"},"path":"some/file/path.txt"},"g":{"grantee":{"Id":{"UserId":{"idp":"0.0.0.0:19000","opaque_id":"f7fbf8c8-139b-4376-b307-cf0a8c2d0d9c","type":1}}},"permissions":{"add_grant":true,"create_container":true,"delete":true,"get_path":true,"get_quota":true,"initiate_file_download":true,"initiate_file_upload":true,"list_grants":true,"list_container":true,"list_file_versions":true,"list_recycle":true,"move":true,"remove_grant":true,"purge_recycle":true,"restore_file_version":true,"restore_recycle_item":true,"stat":true,"update_grant":true,"deny_grant":true}}}`: {200, ``, serverStateEmpty},
	`POST /apps/sciencemesh/~tester/api/DenyGrant {"ref":{"resource_id":{"storage_id":"storage-id","opaque_id":"opaque-id"},"path":"some/file/path.txt"},"g":{"Id":{"UserId":{"idp":"0.0.0.0:19000","opaque_id":"f7fbf8c8-139b-4376-b307-cf0a8c2d0d9c","type":1}}}}`: {200, ``, serverStateEmpty},
	`POST /apps/sciencemesh/~tester/api/RemoveGrant {"ref":{"resource_id":{"storage_id":"storage-id","opaque_id":"opaque-id"},"path":"some/file/path.txt"},"g":{"grantee":{"Id":{"UserId":{"idp":"0.0.0.0:19000","opaque_id":"f7fbf8c8-139b-4376-b307-cf0a8c2d0d9c","type":1}}},"permissions":{"add_grant":true,"create_container":true,"delete":true,"get_path":true,"get_quota":true,"initiate_file_download":true,"initiate_file_upload":true,"list_grants":true,"list_container":true,"list_file_versions":true,"list_recycle":true,"move":true,"remove_grant":true,"purge_recycle":true,"restore_file_version":true,"restore_recycle_item":true,"stat":true,"update_grant":true,"deny_grant":true}}}`: {200, ``, serverStateEmpty},
	`POST /apps/sciencemesh/~tester/api/UpdateGrant {"ref":{"resource_id":{"storage_id":"storage-id","opaque_id":"opaque-id"},"path":"some/file/path.txt"},"g":{"grantee":{"Id":{"UserId":{"idp":"0.0.0.0:19000","opaque_id":"f7fbf8c8-139b-4376-b307-cf0a8c2d0d9c","type":1}}},"permissions":{"add_grant":true,"create_container":true,"delete":true,"get_path":true,"get_quota":true,"initiate_file_download":true,"initiate_file_upload":true,"list_grants":true,"list_container":true,"list_file_versions":true,"list_recycle":true,"move":true,"remove_grant":true,"purge_recycle":true,"restore_file_version":true,"restore_recycle_item":true,"stat":true,"update_grant":true,"deny_grant":true}}}`: {200, ``, serverStateEmpty},
	`POST /apps/sciencemesh/~tester/api/ListGrants {"resource_id":{"storage_id":"storage-id","opaque_id":"opaque-id"},"path":"some/file/path.txt"}`: {200, `[{"grantee":{"type":1,"Id":{"UserId":{"idp":"some-idp","opaque_id":"some-opaque-id","type":1}}},"permissions":{"add_grant":true,"create_container":true,"delete":true,"get_path":true,"get_quota":true,"initiate_file_download":true,"initiate_file_upload":true,"list_grants":true,"list_container":true,"list_file_versions":true,"list_recycle":true,"move":true,"remove_grant":true,"purge_recycle":true,"restore_file_version":true,"restore_recycle_item":true,"stat":true,"update_grant":true,"deny_grant":true}}]`, serverStateEmpty},
	`POST /apps/sciencemesh/~tester/api/GetQuota `: {200, `{"maxBytes":456,"maxFiles":123}`, serverStateEmpty},
	`POST /apps/sciencemesh/~tester/api/CreateReference {"path":"some/file/path.txt","url":"http://bing.com/search?q=dotnet"}`: {200, ``, serverStateEmpty},
	`POST /apps/sciencemesh/~tester/api/Shutdown `: {200, ``, serverStateEmpty},
	`POST /apps/sciencemesh/~tester/api/SetArbitraryMetadata {"ref":{"resource_id":{"storage_id":"storage-id","opaque_id":"opaque-id"},"path":"some/file/path.txt"},"md":{"metadata":{"arbi":"trary","meta":"data"}}}`: {200, ``, serverStateEmpty},
	`POST /apps/sciencemesh/~tester/api/UnsetArbitraryMetadata {"ref":{"resource_id":{"storage_id":"storage-id","opaque_id":"opaque-id"},"path":"some/file/path.txt"},"keys":["arbi"]}`:                                {200, ``, serverStateEmpty},
	`POST /apps/sciencemesh/~tester/api/ListStorageSpaces [{"type":3,"Term":{"Owner":{"idp":"0.0.0.0:19000","opaque_id":"f7fbf8c8-139b-4376-b307-cf0a8c2d0d9c","type":1}}},{"type":2,"Term":{"Id":{"opaque_id":"opaque-id"}}},{"type":4,"Term":{"SpaceType":"home"}}]`: {200, `	[{"opaque":{"map":{"bar":{"value":"c2FtYQ=="},"foo":{"value":"c2FtYQ=="}}},"id":{"opaque_id":"some-opaque-storage-space-id"},"owner":{"id":{"idp":"some-idp","opaque_id":"some-opaque-user-id","type":1}},"root":{"storage_id":"some-storage-ud","opaque_id":"some-opaque-root-id"},"name":"My Storage Space","quota":{"quota_max_bytes":456,"quota_max_files":123},"space_type":"home","mtime":{"seconds":1234567890}}]`, serverStateEmpty},
	`POST /apps/sciencemesh/~tester/api/CreateStorageSpace {"opaque":{"map":{"bar":{"value":"c2FtYQ=="},"foo":{"value":"c2FtYQ=="}}},"owner":{"id":{"idp":"some-idp","opaque_id":"some-opaque-user-id","type":1}},"type":"home","name":"My Storage Space","quota":{"quota_max_bytes":456,"quota_max_files":123}}`: {200, `{"opaque":{"map":{"bar":{"value":"c2FtYQ=="},"foo":{"value":"c2FtYQ=="}}},"owner":{"id":{"idp":"some-idp","opaque_id":"some-opaque-user-id","type":1}},"type":"home","name":"My Storage Space","quota":{"quota_max_bytes":456,"quota_max_files":123}}`, serverStateEmpty},
}

// GetNextcloudServerMock returns a handler that pretends to be a remote Nextcloud server
func GetNextcloudServerMock(called *[]string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(strings.Builder)
		_, err := io.Copy(buf, r.Body)
		if err != nil {
			panic("Error reading response into buffer")
		}
		var key = fmt.Sprintf("%s %s %s", r.Method, r.URL, buf.String())
		fmt.Printf("Nextcloud Server Mock key components %s %d %s %d %s %d\n", r.Method, len(r.Method), r.URL.String(), len(r.URL.String()), buf.String(), len(buf.String()))
		fmt.Printf("Nextcloud Server Mock key %s\n", key)
		*called = append(*called, key)
		response := responses[key]
		if (response == Response{}) {
			key = fmt.Sprintf("%s %s %s %s", r.Method, r.URL, buf.String(), serverState)
			fmt.Printf("Nextcloud Server Mock key with State %s\n", key)
			// *called = append(*called, key)
			response = responses[key]
		}
		if (response == Response{}) {
			fmt.Println("ERROR!!")
			fmt.Println("ERROR!!")
			fmt.Printf("Nextcloud Server Mock key not found! %s\n", key)
			fmt.Println("ERROR!!")
			fmt.Println("ERROR!!")
			response = Response{200, fmt.Sprintf("response not defined! %s", key), serverStateEmpty}
		}
		serverState = responses[key].newServerState
		if serverState == `` {
			serverState = serverStateError
		}
		w.WriteHeader(response.code)
		// w.Header().Set("Etag", "mocker-etag")
		_, err = w.Write([]byte(responses[key].body))
		if err != nil {
			panic(err)
		}
	})
}

// TestingHTTPClient thanks to https://itnext.io/how-to-stub-requests-to-remote-hosts-with-go-6c2c1db32bf2
// Ideally, this function would live in tests/helpers, but
// if we put it there, it gets excluded by .dockerignore, and the
// Docker build fails (see https://github.com/cs3org/reva/issues/1999)
// So putting it here for now - open to suggestions if someone knows
// a better way to inject this.
func TestingHTTPClient(handler http.Handler) (*http.Client, func()) {
	s := httptest.NewServer(handler)

	cli := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, s.Listener.Addr().String())
			},
		},
	}

	return cli, s.Close
}
