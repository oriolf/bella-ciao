package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"testing"
	"time"
)

type testOptions struct {
	method   string
	params   interface{}
	query    string
	token    string
	resToken *string

	file             expectedFile
	fileContent      string
	expectedUsers    []expectedUser
	expectedFiles    []expectedFile
	expectedMessages []string
}

type expectedUser struct {
	uniqueID string
	role     string
	messages []string
}

type expectedFile struct {
	name        string
	description string
}

// TODO error codes should have rational meaning
// TODO HTTP methods should have rational meaning
func TestAPI(t *testing.T) {
	type to = testOptions
	type m = map[string]interface{}
	uniqueID1, uniqueID2, uniqueID3, uniqueID4 := "00000000T", "11111111H", "22222222J", "33333333P"
	user2 := newUser("name", "name@example.com", uniqueID2, "12345678")
	user3 := newUser("name", "name2@example.com", uniqueID3, "12345678")
	user4 := newUser("name", "name@example.com", uniqueID4, "12345678")
	login1 := m{"unique_id": uniqueID1, "password": "12345678"}
	login2 := m{"unique_id": uniqueID2, "password": "12345678"}
	login3 := m{"unique_id": uniqueID3, "password": "12345678"}

	// Initialization and registers

	t.Run("Empty site should not be initialized",
		testEndpoint("/uninitialized", 200, to{}))

	t.Run("Uninitialized site should reject registers",
		testEndpoint("/auth/register", 401, to{method: "POST", params: user2}))

	t.Run("Uninitialized site should reject logins",
		testEndpoint("/auth/login", 401, to{method: "POST", params: login2}))

	admin := newUser("admin", "admin@example.com", uniqueID1, "12345678")
	election := newElection("election", "borda", time.Now().Add(1*time.Hour), time.Now().Add(2*time.Hour), 2, 5)
	t.Run("Empty site can be initialized",
		testEndpoint("/initialize", 200, to{method: "POST", params: m{"admin": admin, "election": election}}))

	t.Run("Initialized site should accept registers",
		testEndpoint("/auth/register", 200, to{method: "POST", params: user2}))

	t.Run("Initialized site should accept registers",
		testEndpoint("/auth/register", 200, to{method: "POST", params: user3}))

	t.Run("Initialized site should not accept registers for duplicate emails",
		testEndpoint("/auth/register", 500, to{method: "POST", params: user4}))

	var token1, token2, token3 string
	t.Run("Initialized site should accept logins from admin",
		testEndpoint("/auth/login", 200, to{method: "POST", params: login1, resToken: &token1}))
	t.Run("Initialized site should accept logins from registered user",
		testEndpoint("/auth/login", 200, to{method: "POST", params: login2, resToken: &token2}))
	t.Run("Initialized site should accept logins from another registered user",
		testEndpoint("/auth/login", 200, to{method: "POST", params: login3, resToken: &token3}))

	t.Run("Check APP State", checkAppState([]expectedUser{
		{uniqueID: uniqueID2, role: ROLE_NONE},
		{uniqueID: uniqueID3, role: ROLE_NONE},
		{uniqueID: uniqueID1, role: ROLE_ADMIN}}))

	// User files management

	t.Run("Non-logged user cannot get files",
		testEndpoint("/users/files/own", 401, to{}))
	t.Run("Non-logged user cannot upload files",
		testEndpoint("/users/files/upload", 401, to{file: expectedFile{description: "file", name: "testfile.txt"}}))
	t.Run("Non-logged user cannot delete files",
		testEndpoint("/users/files/delete", 401, to{query: "?id=1"}))
	t.Run("Non-logged user cannot download files",
		testEndpoint("/users/files/download", 401, to{query: "?id=1"}))

	t.Run("User should not have any files at first",
		testEndpoint("/users/files/own", 200, to{token: token2, expectedFiles: []expectedFile{}}))

	t.Run("User should be able to upload files",
		testEndpoint("/users/files/upload", 200, to{token: token2, file: expectedFile{description: "file", name: "testfile.txt"}}))
	t.Run("User should be able to upload files and get them renamed",
		testEndpoint("/users/files/upload", 200, to{token: token2, file: expectedFile{description: "file", name: "testfile.txt"}}))
	t.Run("User should be able to upload files and get them renamed",
		testEndpoint("/users/files/upload", 200, to{token: token2, file: expectedFile{description: "file", name: "testfile.txt"}}))
	t.Run("Another user should be able to upload files",
		testEndpoint("/users/files/upload", 200, to{token: token3, file: expectedFile{description: "file", name: "testfile.txt"}}))

	t.Run("User should be able to get its files",
		testEndpoint("/users/files/own", 200, to{token: token2, expectedFiles: []expectedFile{
			{name: "testfile.txt", description: "file"}, {name: "testfile_1.txt", description: "file"}, {name: "testfile_2.txt", description: "file"}}}))
	t.Run("User uploaded files should appear in uploads folder", checkUploadsFolder([]string{"testfile.txt", "testfile_1.txt", "testfile_2.txt", "testfile_3.txt"}))

	t.Run("User should be able to delete its files",
		testEndpoint("/users/files/delete", 200, to{token: token2, query: "?id=2"}))
	t.Run("User deleted file should have disappeared",
		testEndpoint("/users/files/own", 200, to{token: token2, expectedFiles: []expectedFile{
			{name: "testfile.txt", description: "file"}, {name: "testfile_2.txt", description: "file"}}}))
	t.Run("User deleted file should have disappeared from uploads folder", checkUploadsFolder([]string{"testfile.txt", "testfile_2.txt", "testfile_3.txt"}))

	t.Run("User should be able to download its files",
		testEndpoint("/users/files/download", 200, to{token: token2, query: "?id=1", fileContent: "file content\n"}))

	t.Run("User should not be able to download deleted files",
		testEndpoint("/users/files/download", 401, to{token: token2, query: "?id=2"}))
	t.Run("User should not be able to download another user's files",
		testEndpoint("/users/files/download", 401, to{token: token2, query: "?id=4"}))
	t.Run("User should not be able to delete another user's files",
		testEndpoint("/users/files/delete", 401, to{token: token2, query: "?id=4"}))

	t.Run("Admin should be able to download another user's files",
		testEndpoint("/users/files/download", 200, to{token: token1, query: "?id=4", fileContent: "file content\n"}))
	t.Run("Admin should be able to delete another user's files",
		testEndpoint("/users/files/delete", 200, to{token: token1, query: "?id=4"}))
	t.Run("User deleted file should have disappeared from uploads folder", checkUploadsFolder([]string{"testfile.txt", "testfile_2.txt"}))

	// User validation, including validation messages

	t.Run("Non-logged user should not get list of unvalidated users",
		testEndpoint("/users/unvalidated/get", 401, to{}))
	t.Run("Non-admin user should not get list of unvalidated users",
		testEndpoint("/users/unvalidated/get", 401, to{token: token2}))
	t.Run("Admin user should get list of unvalidated users",
		testEndpoint("/users/unvalidated/get", 200, to{token: token1, expectedUsers: []expectedUser{
			{uniqueID: uniqueID2}, {uniqueID: uniqueID3}}}))

	t.Run("Non-logged user should not be able to add messages",
		testEndpoint("/users/messages/add", 401, to{params: m{"user_id": 2, "content": "message content"}}))
	t.Run("Non-admin user should not be able to add messages",
		testEndpoint("/users/messages/add", 401, to{token: token2, params: m{"user_id": 2, "content": "message content"}}))
	t.Run("Admin user should be able to add messages",
		testEndpoint("/users/messages/add", 200, to{token: token1, params: m{"user_id": 2, "content": "message content user 2"}}))
	t.Run("Admin user should be able to add messages",
		testEndpoint("/users/messages/add", 200, to{token: token1, params: m{"user_id": 2, "content": "message content user 2"}}))
	t.Run("Admin user should be able to add messages",
		testEndpoint("/users/messages/add", 200, to{token: token1, params: m{"user_id": 3, "content": "message content user 3"}}))

	t.Run("List of unvalidated users should contain messages",
		testEndpoint("/users/unvalidated/get", 200, to{token: token1, expectedUsers: []expectedUser{
			{uniqueID: uniqueID2, messages: []string{"message content user 2", "message content user 2"}}, {uniqueID: uniqueID3, messages: []string{"message content user 3"}}}}))

	t.Run("Non-logged user should not be able to get messages",
		testEndpoint("/users/messages/own", 401, to{}))
	t.Run("Logged user should get its own messages",
		testEndpoint("/users/messages/own", 200, to{token: token2, expectedMessages: []string{"message content user 2", "message content user 2"}}))

	// "Non-logged user should not be able to solve messages"
	// "Logged user should be able to solve its messages"
	// "Logged user should not be able to solve another user's messages"
	// "Admin user should be able to solve another user's messages"
	// check again which messages remain

	// Non-logged user should not be able to validate users
	// Non-admin user should not be able to validate users
	// Admin user should be able to validate users
	// check app state again to see the validated user
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

		req, err := http.NewRequest(method, path+options.query, body)
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

		if options.expectedMessages != nil {
			var messages []UserMessage
			if err := json.Unmarshal([]byte(rr.Body.String()), &messages); err != nil {
				t.Errorf("Could not unmarshal expected messages response: %s", err)
			} else {
				compareMessages(t, options.expectedMessages, messages)
			}
		}

		if options.fileContent != "" && options.fileContent != rr.Body.String() {
			t.Errorf("Wrong file contents. Expected %q but found %q.", options.fileContent, rr.Body.String())
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
				compareMessages(t, e.messages, u.Messages)
				continue LOOP
			}
		}
		t.Errorf("Expected user with unique ID %q, but none found.", e.uniqueID)
	}
}

func compareMessages(t *testing.T, expected []string, got []UserMessage) {
	if len(expected) != len(got) {
		t.Errorf("Expected %d messages, but got %d.", len(expected), len(got))
		return
	}

	var gotStrings []string
	for _, x := range got {
		gotStrings = append(gotStrings, x.Content)
	}
	sort.Strings(expected)
	sort.Strings(gotStrings)
	for i := range expected {
		if expected[i] != gotStrings[i] {
			t.Errorf("Expected %d message to be %q, but got %q.", i, expected[i], gotStrings[i])
		}
	}
}

func checkUploadsFolder(expectedFiles []string) func(*testing.T) {
	return func(t *testing.T) {
		files, err := ioutil.ReadDir("uploads")
		if err != nil {
			t.Errorf("Could not read uploads dir: %s.", err)
			return
		}

		if len(files) != len(expectedFiles) {
			t.Errorf("Expected %d files, but got %d.", len(expectedFiles), len(files))
		}

	LOOP:
		for _, e := range expectedFiles {
			for _, f := range files {
				if e == f.Name() {
					continue LOOP
				}
			}
			t.Errorf("Expected file in uploads with name %s, but found none", e)
		}
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
