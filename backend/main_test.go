package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

type testOptions struct {
	method   string
	params   interface{}
	token    string
	resToken *string
	file     expectedFile

	expectedUsers []expectedUser
	expectedFiles []expectedFile
}

type expectedUser struct {
	uniqueID string
	role     string
}

type expectedFile struct {
	name        string
	description string
}

func TestAPI(t *testing.T) {
	type to = testOptions
	type m = map[string]interface{}
	uniqueID0, uniqueID1, uniqueID2, uniqueID3 := "00000000T", "11111111H", "22222222J", "33333333P"
	user1 := newUser("name", "name@example.com", uniqueID1, "12345678")
	user2 := newUser("name", "name2@example.com", uniqueID2, "12345678")
	user3 := newUser("name", "name@example.com", uniqueID3, "12345678")
	login0 := m{"unique_id": uniqueID0, "password": "12345678"}
	login1 := m{"unique_id": uniqueID1, "password": "12345678"}

	t.Run("Empty site should not be initialized",
		testEndpoint("/uninitialized", 200, to{}))

	t.Run("Uninitialized site should reject registers",
		testEndpoint("/auth/register", 401, to{method: "POST", params: user1}))

	t.Run("Uninitialized site should reject logins",
		testEndpoint("/auth/login", 401, to{method: "POST", params: login1}))

	admin := newUser("admin", "admin@example.com", uniqueID0, "12345678")
	election := newElection("election", "borda", time.Now().Add(1*time.Hour), time.Now().Add(2*time.Hour), 2, 5)
	t.Run("Empty site can be initialized",
		testEndpoint("/initialize", 200, to{method: "POST", params: m{"admin": admin, "election": election}}))

	t.Run("Initialized site should accept registers",
		testEndpoint("/auth/register", 200, to{method: "POST", params: user1}))

	t.Run("Initialized site should accept registers",
		testEndpoint("/auth/register", 200, to{method: "POST", params: user2}))

	t.Run("Initialized site should not accept registers for duplicate emails",
		testEndpoint("/auth/register", 500, to{method: "POST", params: user3}))

	var token0, token1 string
	t.Run("Initialized site should accept logins from admin",
		testEndpoint("/auth/login", 200, to{method: "POST", params: login0, resToken: &token0}))
	t.Run("Initialized site should accept logins from registered users",
		testEndpoint("/auth/login", 200, to{method: "POST", params: login1, resToken: &token1}))

	t.Run("Check APP State", checkAppState([]expectedUser{
		{uniqueID: uniqueID1, role: ROLE_NONE},
		{uniqueID: uniqueID2, role: ROLE_NONE},
		{uniqueID: uniqueID0, role: ROLE_ADMIN}}))

	t.Run("Non-logged user should not get list of unvalidated users",
		testEndpoint("/users/unvalidated/get", 401, to{}))

	t.Run("Non-admin user should not get list of unvalidated users",
		testEndpoint("/users/unvalidated/get", 401, to{token: token1}))

	t.Run("Admin user should get list of unvalidated users",
		testEndpoint("/users/unvalidated/get", 200, to{token: token0, expectedUsers: []expectedUser{
			{uniqueID: uniqueID1}, {uniqueID: uniqueID2}}}))

	t.Run("Non-logged user cannot get files",
		testEndpoint("/users/files/own", 401, to{}))
	// TODO "Non-logged user cannot upload files"
	// TODO "Non-logged user cannot delete files"
	// TODO "Non-logged user cannot download files"

	t.Run("User should not have any files at first",
		testEndpoint("/users/files/own", 200, to{token: token1, expectedFiles: []expectedFile{}}))

	t.Run("User should be able to upload files",
		testEndpoint("/users/files/upload", 200, to{token: token1, file: expectedFile{description: "file", name: "testfile.txt"}}))
	t.Run("User should be able to upload files and get them renamed",
		testEndpoint("/users/files/upload", 200, to{token: token1, file: expectedFile{description: "file", name: "testfile.txt"}}))
	t.Run("User should be able to upload files and get them renamed",
		testEndpoint("/users/files/upload", 200, to{token: token1, file: expectedFile{description: "file", name: "testfile.txt"}}))
	t.Run("User should be able to get its files",
		testEndpoint("/users/files/own", 200, to{token: token1, expectedFiles: []expectedFile{
			{name: "testfile.txt", description: "file"}, {name: "testfile_1.txt", description: "file"}, {name: "testfile_2.txt", description: "file"},
		}}))
	// TODO "User should be able to download its files"
	// TODO "User should be able to delete its files"
	// TODO "User should not be able to download another user files"
	// TODO "User should not be able to delete another user files"
	// TODO "Admin should be able to download another user files"
	// TODO "Admin should be able to delete another user files"
}

func newUser(name, email, uniqueID, password string) map[string]interface{} {
	return map[string]interface{}{
		"name":      name,
		"email":     email,
		"unique_id": uniqueID,
		"password":  password,
	}
}

func newElection(name, countType string, start, end time.Time, minCandidates, maxCandidates int) map[string]interface{} {
	return map[string]interface{}{
		"name":           name,
		"start":          start.Format(time.RFC3339Nano),
		"end":            end.Format(time.RFC3339Nano),
		"count_type":     countType,
		"min_candidates": minCandidates,
		"max_candidates": maxCandidates,
	}
}

func testEndpoint(path string, expectedCode int, options testOptions) func(*testing.T) {
	i := 0
	return func(t *testing.T) {
		i++
		method := "GET"
		var params interface{}
		if options.method != "" {
			method = options.method
		}

		if options.params != nil {
			params = options.params
		}

		var body io.Reader
		var err error
		contentType := ""
		if params != nil {
			b, err := json.Marshal(params)
			if err != nil {
				t.Fatalf("[%d] Could not marshal params for endpoint %q. Error: %s", i, path, err)
			}
			body = bytes.NewReader(b)
			contentType = "application/json"
		} else if options.file.name != "" {
			body, contentType, err = fileUploadBody(options.file.name, map[string]string{"description": options.file.description})
			if err != nil {
				t.Fatalf("[%d] Could not create file upload body for endpoint %q. Error: %s\n", i, path, err)
			}
		}

		req, err := http.NewRequest(method, path, body)
		if err != nil {
			t.Fatalf("[%d] Could not create request for endpoint %q. Error: %s\n", i, path, err)
		}

		if options.token != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", options.token))
		}
		req.Header.Set("Content-Type", contentType)

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(appHandlers[path])
		handler.ServeHTTP(rr, req)
		if options.resToken != nil {
			token := rr.Body.String()
			*options.resToken = strings.Trim(token, "\"")
		}
		if rr.Code != expectedCode {
			t.Errorf("[%d] Expected code %v testing endpoint %q, but got %v.", i, expectedCode, path, rr.Code)
		}
		if options.expectedUsers != nil {
			var users []User
			if err := json.Unmarshal([]byte(rr.Body.String()), &users); err != nil {
				t.Errorf("Could not unmarshal expected users response: %s", err)
			} else {
				compareUsers(t, options.expectedUsers, users)
			}
		}

		if options.expectedFiles != nil {
			var files []UserFile
			if err := json.Unmarshal([]byte(rr.Body.String()), &files); err != nil {
				t.Errorf("Could not unmarshal expected files response: %s", err)
			} else {
				compareFiles(t, options.expectedFiles, files)
			}
		}
	}
}

func checkAppState(expectedUsers []expectedUser) func(*testing.T) {
	return func(t *testing.T) {
		db, err := sql.Open("sqlite3", dbfile)
		if err != nil {
			t.Fatal("Error during database connection in handler:", err)
		}
		defer db.Close()

		users, err := getAllUsers(db)
		if err != nil {
			t.Errorf("Error getting all users: %s.", err)
		}

		compareUsers(t, expectedUsers, users)
	}
}

func compareUsers(t *testing.T, expectedUsers []expectedUser, users []User) {
	if len(users) != len(expectedUsers) {
		t.Errorf("Expected %d users, but got %d.", len(expectedUsers), len(users))
	}

LOOP:
	for _, e := range expectedUsers {
		for _, u := range users {
			if e.uniqueID == u.UniqueID {
				if e.role != u.Role {
					t.Errorf("Expected user with unique ID %q to have role %q, but has role %q.", e.uniqueID, e.role, u.Role)
				}
				continue LOOP
			}
		}
		t.Errorf("Expected user with unique ID %q, but none found.", e.uniqueID)
	}
}

func compareFiles(t *testing.T, expectedFiles []expectedFile, files []UserFile) {
	if len(files) != len(expectedFiles) {
		t.Errorf("Expected %d files, but got %d.", len(expectedFiles), len(files))
	}

LOOP:
	for _, e := range expectedFiles {
		for _, u := range files {
			if e.name == u.Name {
				if e.description != u.Description {
					t.Errorf("Expected file with name %q to have description %q, but has description %q.", e.name, e.description, u.Description)
				}
				continue LOOP
			}
		}
		t.Errorf("Expected file with name %q, but none found.", e.name)
	}
}

func fileUploadBody(filename string, params map[string]string) (io.Reader, string, error) {
	file, err := os.Open("../test/" + filename)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, "", err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, "", err
	}

	for key, val := range params {
		if err = writer.WriteField(key, val); err != nil {
			return nil, "", err
		}
	}

	if err = writer.Close(); err != nil {
		return nil, "", err
	}

	return body, writer.FormDataContentType(), err
}
