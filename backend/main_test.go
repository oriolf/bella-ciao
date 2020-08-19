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

	"github.com/google/go-cmp/cmp"
)

type testOptions struct {
	method    string
	params    interface{}
	query     string
	token     string
	resToken  *string
	candidate Candidate

	file                     expectedFile
	fileContent              string
	expectedUsers            []expectedUser
	expectedFiles            []expectedFile
	expectedUnsolvedMessages []string
	expectedCandidates       []Candidate
	expectedElections        []Election
}

type expectedUser struct {
	uniqueID         string
	role             string
	unsolvedMessages []string
}

type expectedFile struct {
	name        string
	description string
}

// TODO HTTP methods should have rational meaning
func TestAPI(t *testing.T) {
	bootstrap()

	type to = testOptions
	type m = map[string]interface{}
	uniqueID1, uniqueID2, uniqueID3, uniqueID4 := "11111111H", "22222222J", "33333333P", "44444444A"
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

	admin := newUser("admin", "admin@example.com", "21111111H", "12345678")
	appConfig := m{"id_formats": []string{ID_DNI}}
	electionStart, electionEnd := time.Now().Add(1*time.Hour), time.Now().Add(2*time.Hour)

	// wrong admin unique id
	election := newElection("election", COUNT_BORDA, electionStart, electionEnd, 2, 5)
	t.Run("Empty site cannot be initialized with wrong parameters",
		testEndpoint("/initialize", 400, to{method: "POST", params: m{"admin": admin, "election": election, "config": appConfig}}))

	// wrong election start and end
	admin["unique_id"] = uniqueID1
	election = newElection("election", COUNT_BORDA, electionEnd, electionStart, 2, 5)
	t.Run("Empty site cannot be initialized with wrong parameters",
		testEndpoint("/initialize", 400, to{method: "POST", params: m{"admin": admin, "election": election, "config": appConfig}}))

	// wrong election min and max candidates
	election = newElection("election", COUNT_BORDA, electionStart, electionEnd, 5, 2)
	t.Run("Empty site cannot be initialized with wrong parameters",
		testEndpoint("/initialize", 400, to{method: "POST", params: m{"admin": admin, "election": election, "config": appConfig}}))

	// wrong count method
	election = newElection("election", "invalid_count", electionStart, electionEnd, 2, 5)
	t.Run("Empty site cannot be initialized with wrong parameters",
		testEndpoint("/initialize", 400, to{method: "POST", params: m{"admin": admin, "election": election, "config": appConfig}}))

	// empty id formats list
	election = newElection("election", COUNT_BORDA, electionStart, electionEnd, 2, 5)
	t.Run("Empty site cannot be initialized with wrong parameters",
		testEndpoint("/initialize", 400, to{method: "POST", params: m{"admin": admin, "election": election, "config": m{"id_formats": []string{}}}}))

	t.Run("Empty site can be initialized",
		testEndpoint("/initialize", 200, to{method: "POST", params: m{"admin": admin, "election": election, "config": appConfig}}))

	t.Run("Users should be able to register",
		testEndpoint("/auth/register", 200, to{method: "POST", params: user2}))
	t.Run("Users should be able to register",
		testEndpoint("/auth/register", 200, to{method: "POST", params: user3}))
	t.Run("Registers with invalid parameters should be rejected",
		testEndpoint("/auth/register", 400, to{method: "POST", params: m{"email": "asd@example.com"}}))
	t.Run("Registers with duplicated emails should be rejected",
		testEndpoint("/auth/register", 500, to{method: "POST", params: user4}))

	var token1, token2, token3 string
	t.Run("Admin should be able to log in",
		testEndpoint("/auth/login", 200, to{method: "POST", params: login1, resToken: &token1}))
	t.Run("User should be able to log in",
		testEndpoint("/auth/login", 200, to{method: "POST", params: login2, resToken: &token2}))
	t.Run("Another user should be able to log in",
		testEndpoint("/auth/login", 200, to{method: "POST", params: login3, resToken: &token3}))
	t.Run("Log in with invalid parameters should be rejected",
		testEndpoint("/auth/login", 400, to{method: "POST", params: m{"unique_id": uniqueID1}}))

	t.Run("Check APP State", checkAppState([]expectedUser{
		{uniqueID: uniqueID2, role: ROLE_NONE},
		{uniqueID: uniqueID3, role: ROLE_NONE},
		{uniqueID: uniqueID1, role: ROLE_ADMIN}}))

	// TODO when id formats are checked
	//	userNIE := newUser("name", "asd3@example.com", "X1111111G", "12345678")
	//	t.Run("Registers with invalid id format should be rejected",
	//		testEndpoint("/auth/register", 500, to{method: "POST", params: userNIE}))

	t.Run("Admin can update site config",
		testEndpoint("/config/update", 200, to{method: "POST", token: token1, params: m{"id_formats": []string{ID_DNI, ID_NIE}}}))
	t.Run("Admin cannot remove id formats from site config",
		testEndpoint("/config/update", 500, to{method: "POST", token: token1, params: m{"id_formats": []string{ID_DNI}}}))

	// TODO test userNIE can register now

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
	t.Run("User should not be able to upload files without description",
		testEndpoint("/users/files/upload", 400, to{token: token2, file: expectedFile{description: "", name: "testfile.txt"}}))

	t.Run("User should be able to get its files",
		testEndpoint("/users/files/own", 200, to{token: token2, expectedFiles: []expectedFile{
			{name: "testfile.txt", description: "file"}, {name: "testfile_1.txt", description: "file"}, {name: "testfile_2.txt", description: "file"}}}))
	t.Run("User uploaded files should appear in uploads folder", checkUploadsFolder([]string{"testfile.txt", "testfile_1.txt", "testfile_2.txt", "testfile_3.txt"}))

	t.Run("User should be able to delete its files",
		testEndpoint("/users/files/delete", 200, to{token: token2, query: "?id=2"}))
	t.Run("User should not be able to delete files with invalid parameters",
		testEndpoint("/users/files/delete", 400, to{token: token2, query: "?id=-1"}))

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
	t.Run("User should not be able to download files with invalid parameters",
		testEndpoint("/users/files/download", 400, to{token: token2, query: "?id=0"}))
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
			{uniqueID: uniqueID2, unsolvedMessages: []string{"message content user 2", "message content user 2"}},
			{uniqueID: uniqueID3, unsolvedMessages: []string{"message content user 3"}}}}))

	t.Run("Non-logged user should not be able to get messages",
		testEndpoint("/users/messages/own", 401, to{}))
	t.Run("Logged user should get its own messages",
		testEndpoint("/users/messages/own", 200, to{token: token2, expectedUnsolvedMessages: []string{"message content user 2", "message content user 2"}}))

	t.Run("Non-logged user should not be able to solve messages",
		testEndpoint("/users/messages/solve", 401, to{query: "?id=1"}))
	t.Run("Logged user should be able to solve its messages",
		testEndpoint("/users/messages/solve", 200, to{token: token2, query: "?id=1"}))
	t.Run("Logged user should not be able to solve another user's messages",
		testEndpoint("/users/messages/solve", 401, to{token: token2, query: "?id=3"}))
	t.Run("Logged user should not be able to solve messages with invalid parameters",
		testEndpoint("/users/messages/solve", 400, to{token: token2, query: "?id=-123"}))
	t.Run("Admin user should be able to solve another user's messages",
		testEndpoint("/users/messages/solve", 200, to{token: token1, query: "?id=3"}))

	t.Run("Validated messages should not appear",
		testEndpoint("/users/messages/own", 200, to{token: token2, expectedUnsolvedMessages: []string{"message content user 2"}}))
	t.Run("Validated messages should not appear",
		testEndpoint("/users/messages/own", 200, to{token: token3, expectedUnsolvedMessages: []string{}}))

	t.Run("The only validated user should be admin",
		testEndpoint("/users/validated/get", 200, to{token: token1, expectedUsers: []expectedUser{{uniqueID: uniqueID1}}}))

	t.Run("Non-logged user should not be able to validate users",
		testEndpoint("/users/validate", 401, to{query: "?id=2"}))
	t.Run("Non-admin user should not be able to validate users",
		testEndpoint("/users/validate", 401, to{token: token2, query: "?id=2"}))
	t.Run("Admin user should be able to validate users with invalid paramters",
		testEndpoint("/users/validate", 400, to{token: token1, query: "?id=0"}))
	t.Run("Admin user should be able to validate users",
		testEndpoint("/users/validate", 200, to{token: token1, query: "?id=2"}))

	t.Run("Unexisting users cannot be validated",
		testEndpoint("/users/validate", 500, to{token: token1, query: "?id=999"}))
	t.Run("Validated users cannot be validated again",
		testEndpoint("/users/validate", 500, to{token: token1, query: "?id=2"}))
	t.Run("Admin users cannot be validated",
		testEndpoint("/users/validate", 500, to{token: token1, query: "?id=1"}))

	t.Run("Check APP State", checkAppState([]expectedUser{
		{uniqueID: uniqueID3, role: ROLE_NONE},
		{uniqueID: uniqueID2, role: ROLE_VALIDATED},
		{uniqueID: uniqueID1, role: ROLE_ADMIN}}))

	t.Run("There should be two validated users",
		testEndpoint("/users/validated/get", 200, to{token: token1, expectedUsers: []expectedUser{
			{uniqueID: uniqueID1}, {uniqueID: uniqueID2, unsolvedMessages: []string{"message content user 2"}}}}))

	candidate1 := Candidate{Name: "candidate 1", Presentation: "candidate 1 presentation", Image: "candidate.jpg"}
	candidate2 := Candidate{Name: "candidate 2", Presentation: "candidate 2 presentation", Image: "candidate.jpg"}
	candidateWrong := Candidate{Name: "", Presentation: "candidate wrong presentation", Image: "candidate.jpg"}

	t.Run("Non-logged users should not be able to add candidates",
		testEndpoint("/candidates/add", 401, to{candidate: candidate1}))
	t.Run("Non-admin users should not be able to add candidates",
		testEndpoint("/candidates/add", 401, to{token: token2, candidate: candidate1}))
	t.Run("Admin users should not be able to add candidates with invalid parameters",
		testEndpoint("/candidates/add", 400, to{token: token1, candidate: candidateWrong}))
	t.Run("Admin users should be able to add candidates",
		testEndpoint("/candidates/add", 200, to{token: token1, candidate: candidate1}))
	t.Run("Admin users should be able to add candidates",
		testEndpoint("/candidates/add", 200, to{token: token1, candidate: candidate2}))

	candidate2.Image = "candidate_1.jpg"
	t.Run("Non-logged users should be able to get candidates",
		testEndpoint("/candidates/get", 200, to{expectedCandidates: []Candidate{candidate1, candidate2}}))
	t.Run("Logged users should be able to get candidates",
		testEndpoint("/candidates/get", 200, to{token: token2, expectedCandidates: []Candidate{candidate1, candidate2}}))
	t.Run("Candidate images should appear in uploads folder", checkUploadsFolder([]string{"testfile.txt", "testfile_2.txt", "candidate.jpg", "candidate_1.jpg"}))

	t.Run("Non-logged users should not be able to delete candidates",
		testEndpoint("/candidates/delete", 401, to{query: "?id=2"}))
	t.Run("Non-admin users should not be able to delete candidates",
		testEndpoint("/candidates/delete", 401, to{token: token2, query: "?id=2"}))
	t.Run("Admin users should be able to delete candidates",
		testEndpoint("/candidates/delete", 200, to{token: token1, query: "?id=2"}))

	t.Run("Deleted candidate should not appear anymore",
		testEndpoint("/candidates/get", 200, to{expectedCandidates: []Candidate{candidate1}}))
	t.Run("Deleted candidate image should not appear anymore", checkUploadsFolder([]string{"testfile.txt", "testfile_2.txt", "candidate.jpg"}))

	candidate1.ID = 1
	candidate1.ElectionID = 1
	election.ID = 1
	election.Candidates = []Candidate{candidate1}
	t.Run("Non-logged user should be able to see any elections yet",
		testEndpoint("/elections/get", 200, to{expectedElections: []Election{}}))
	t.Run("Non-admin user should be able to see any elections yet",
		testEndpoint("/elections/get", 200, to{token: token2, expectedElections: []Election{}}))
	t.Run("Admin user should be able to see unpublished elections",
		testEndpoint("/elections/get", 200, to{token: token1, expectedElections: []Election{election}}))

	t.Run("Non-logged user should not be able to publish election",
		testEndpoint("/elections/publish", 401, to{query: "?id=1"}))
	t.Run("Non-admin user should not be able to publish election",
		testEndpoint("/elections/publish", 401, to{token: token2, query: "?id=1"}))
	t.Run("Admin user should be able to publish election",
		testEndpoint("/elections/publish", 200, to{token: token1, query: "?id=1"}))

	election.Public = true
	t.Run("Non-logged user should be able to see elections",
		testEndpoint("/elections/get", 200, to{expectedElections: []Election{election}}))
}

func newUser(name, email, uniqueID, password string) map[string]interface{} {
	return map[string]interface{}{
		"name":      name,
		"email":     email,
		"unique_id": uniqueID,
		"password":  password,
	}
}

func newElection(name, countMethod string, start, end time.Time, minCandidates, maxCandidates int) Election {
	return Election{
		Name:          name,
		Start:         start,
		End:           end,
		CountMethod:   countMethod,
		MinCandidates: minCandidates,
		MaxCandidates: maxCandidates,
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
			body, contentType, err = fileUploadBody(options.file.name, "file", map[string]string{"description": options.file.description})
			if err != nil {
				t.Fatalf("[%d] Could not create file upload body for endpoint %q. Error: %s\n", i, path, err)
			}
		} else if options.candidate.Name != "" {
			body, contentType, err = fileUploadBody(options.candidate.Image, "image", map[string]string{
				"name":         options.candidate.Name,
				"presentation": options.candidate.Presentation,
			})
			if err != nil {
				t.Fatalf("[%d] Could not create candidate file upload body for endpoint %q. Error: %s\n", i, path, err)
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

		if options.expectedCandidates != nil {
			var candidates []Candidate
			if err := json.Unmarshal([]byte(rr.Body.String()), &candidates); err != nil {
				t.Errorf("Could not unmarshal expected candidates response: %s", err)
			} else {
				compareCandidates(t, options.expectedCandidates, candidates)
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

		if options.expectedUnsolvedMessages != nil {
			var messages []UserMessage
			if err := json.Unmarshal([]byte(rr.Body.String()), &messages); err != nil {
				t.Errorf("Could not unmarshal expected messages response: %s", err)
			} else {
				compareMessages(t, options.expectedUnsolvedMessages, messages)
			}
		}

		if options.expectedElections != nil {
			var elections []Election
			if err := json.Unmarshal([]byte(rr.Body.String()), &elections); err != nil {
				t.Errorf("Could not unmarshal expected messages response: %s", err)
			} else {
				compareElections(t, options.expectedElections, elections)
			}
		}

		if options.fileContent != "" && options.fileContent != rr.Body.String() {
			t.Errorf("Wrong file contents. Expected %q but found %q.", options.fileContent, rr.Body.String())
		}
	}
}

func checkAppState(expectedUsers []expectedUser) func(*testing.T) {
	return func(t *testing.T) {
		db, err := sql.Open("sqlite3", DB_FILE)
		if err != nil {
			t.Fatal("Error during database connection in handler:", err)
		}
		defer db.Close()

		tx, _ := db.Begin()
		users, err := getAllUsers(tx)
		if err != nil {
			t.Errorf("Error getting all users: %s.", err)
		}
		tx.Rollback()

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
				compareMessages(t, e.unsolvedMessages, u.Messages)
				continue LOOP
			}
		}
		t.Errorf("Expected user with unique ID %q, but none found.", e.uniqueID)
	}
}

func compareCandidates(t *testing.T, expected, got []Candidate) {
	if len(expected) != len(got) {
		t.Errorf("Expected %d candidates, but got %d.", len(expected), len(got))
		return
	}

	sort.Slice(expected, func(i, j int) bool { return expected[i].Name < expected[j].Name })
	sort.Slice(got, func(i, j int) bool { return got[i].Name < got[j].Name })
	for i := range expected {
		if !equalCandidates(expected[i], got[i]) {
			t.Errorf("Expected user %v but got %v.", expected[i], got[i])
		}
	}
}

func equalCandidates(a, b Candidate) bool {
	return a.Name == b.Name && a.Presentation == b.Presentation && a.Image == b.Image
}

func compareMessages(t *testing.T, expected []string, got []UserMessage) {
	var gotStrings []string
	for _, x := range got {
		if !x.Solved {
			gotStrings = append(gotStrings, x.Content)
		}
	}

	if len(expected) != len(gotStrings) {
		t.Errorf("Expected %d messages, but got %d.", len(expected), len(gotStrings))
		return
	}

	sort.Strings(expected)
	sort.Strings(gotStrings)
	for i := range expected {
		if expected[i] != gotStrings[i] {
			t.Errorf("Expected %d message to be %q, but got %q.", i, expected[i], gotStrings[i])
		}
	}
}

func compareElections(t *testing.T, expected []Election, got []Election) {
	if len(expected) != len(got) {
		t.Errorf("Expected %d elections, but got %d", len(expected), len(got))
		return
	}

	if len(expected) > 0 {
		if diff := cmp.Diff(expected, got); diff != "" {
			t.Errorf("Expected no diff in elections, but got: %s.", diff)
		}
	}
}

func checkUploadsFolder(expectedFiles []string) func(*testing.T) {
	return func(t *testing.T) {
		files, err := ioutil.ReadDir(UPLOADS_FOLDER)
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

func fileUploadBody(filename string, fileField string, params map[string]string) (io.Reader, string, error) {
	file, err := os.Open("../test/" + filename)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(fileField, filename)
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

type testValidateID struct {
	s             string
	expectedError bool
}

func TestValidateDNI(t *testing.T) {
	for i, test := range []testValidateID{
		{s: "11111111H", expectedError: false},
		{s: "22222222J", expectedError: false},
		{s: "33333333P", expectedError: false},
		{s: "44444444A", expectedError: false},
		{s: "12345678Z", expectedError: false},

		{s: "", expectedError: true},
		{s: "11111111 H", expectedError: true},
		{s: "11111111", expectedError: true},
		{s: "22222222H", expectedError: true},
		{s: "11111111h", expectedError: true},
	} {
		if err := validateDNI(test.s); err != nil && !test.expectedError {
			t.Errorf("[%d] Expected no error but got %q.", i, err)
		} else if err == nil && test.expectedError {
			t.Errorf("[%d] Expected an error but got none.", i)
		}
	}
}

func TestValidateNIE(t *testing.T) {
	for i, test := range []testValidateID{
		{s: "X1111111G", expectedError: false},
		{s: "Y2222222E", expectedError: false},

		{s: "", expectedError: true},
		{s: "X111111G", expectedError: true},
		{s: "X 1111111 G", expectedError: true},
		{s: "X1111111A", expectedError: true},
	} {
		if err := validateNIE(test.s); err != nil && !test.expectedError {
			t.Errorf("[%d] Expected no error but got %q.", i, err)
		} else if err == nil && test.expectedError {
			t.Errorf("[%d] Expected an error but got none.", i)
		}
	}
}

func TestValidatePassport(t *testing.T) {
	for i, test := range []testValidateID{
		{s: "ABC123456A", expectedError: false},
		{s: "XYZ111111B", expectedError: false},
		{s: "XYZ111111C", expectedError: false},

		{s: "", expectedError: true},
		{s: "ABC 123456 A", expectedError: true},
		{s: "ABC-123456-A", expectedError: true},
		{s: "ABC12345B", expectedError: true},
	} {
		if err := validatePassport(test.s); err != nil && !test.expectedError {
			t.Errorf("[%d] Expected no error but got %q.", i, err)
		} else if err == nil && test.expectedError {
			t.Errorf("[%d] Expected an error but got none.", i)
		}
	}
}
