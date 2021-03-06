package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
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
	method     string
	params     interface{}
	postParams map[string]string
	query      string
	cookies    []*http.Cookie
	resCookies *[]*http.Cookie
	candidate  Candidate
	voteToken  *string

	file                     expectedFile
	fileContent              string
	expectedUser             expectedUser
	expectedUsers            expectedUsersResponse
	expectedFiles            []expectedFile
	expectedUnsolvedMessages []string
	expectedCandidates       []Candidate
	expectedElections        []Election
}

type expectedUsersResponse struct {
	Users []expectedUser
	Total int
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
// TODO test concurrent queries that insert in the database all work properly
// TODO test that endpoints used in frontend match endpoints defined in appHandlers
// TODO test user pagination properly
// TODO load test
func TestAPI(t *testing.T) {
	bootstrap()
	globalTesting = true

	type to = testOptions
	type m = map[string]interface{}
	type ms = map[string]string
	uniqueID1, uniqueID2, uniqueID3, uniqueID4, uniqueID5 := "11111111H", "22222222J", "33333333P", "44444444A", "X1111111G"
	user2 := newUser("User 2 name", "name@example.com", uniqueID2, "12345678")
	user3 := newUser("User 3 name", "name2@example.com", uniqueID3, "12345678")
	user4 := newUser("User 4 name", "name@example.com", uniqueID4, "12345678")
	login1 := m{"unique_id": uniqueID1, "password": "12345678"}
	login2 := m{"unique_id": uniqueID2, "password": "12345678"}
	login3 := m{"unique_id": uniqueID3, "password": "12345678"}
	login5 := m{"unique_id": uniqueID5, "password": "12345678"}

	// Initialization and registers

	t.Run("Empty site should not be initialized",
		testEndpoint("/uninitialized", 200, to{}))

	t.Run("Uninitialized site should reject registers",
		testEndpoint("/auth/register", 401, to{method: "POST", params: user2}))

	t.Run("Uninitialized site should reject logins",
		testEndpoint("/auth/login", 401, to{method: "POST", params: login2}))

	admin := newUser("admin", "admin@example.com", "21111111H", "12345678")
	appConfig := m{"id_formats": []string{ID_DNI}}
	electionStart, electionEnd := now().Add(1*time.Hour), now().Add(2*time.Hour)

	// wrong admin unique id
	election := newElection("election", COUNT_BORDA, electionStart, electionEnd, 2, 3)
	t.Run("Empty site cannot be initialized with wrong parameters",
		testEndpoint("/initialize", 400, to{method: "POST", params: m{"admin": admin, "election": election, "config": appConfig}}))

	// wrong election start and end
	admin["unique_id"] = uniqueID1
	election = newElection("election", COUNT_BORDA, electionEnd, electionStart, 2, 3)
	t.Run("Empty site cannot be initialized with wrong parameters",
		testEndpoint("/initialize", 400, to{method: "POST", params: m{"admin": admin, "election": election, "config": appConfig}}))

	// wrong election min and max candidates
	election = newElection("election", COUNT_BORDA, electionStart, electionEnd, 3, 2)
	t.Run("Empty site cannot be initialized with wrong parameters",
		testEndpoint("/initialize", 400, to{method: "POST", params: m{"admin": admin, "election": election, "config": appConfig}}))

	// wrong count method
	election = newElection("election", "invalid_count", electionStart, electionEnd, 2, 3)
	t.Run("Empty site cannot be initialized with wrong parameters",
		testEndpoint("/initialize", 400, to{method: "POST", params: m{"admin": admin, "election": election, "config": appConfig}}))

	// empty id formats list
	election = newElection("election", COUNT_BORDA, electionStart, electionEnd, 2, 3)
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

	var cookies1, cookies2, cookies3, cookies5, cookieslogout []*http.Cookie
	t.Run("Admin should be able to log in",
		testEndpoint("/auth/login", 200, to{method: "POST", params: login1, resCookies: &cookies1}))
	t.Run("User should be able to log in",
		testEndpoint("/auth/login", 200, to{method: "POST", params: login2, resCookies: &cookies2}))
	t.Run("Another user should be able to log in",
		testEndpoint("/auth/login", 200, to{method: "POST", params: login3, resCookies: &cookies3}))
	t.Run("Log in with invalid parameters should be rejected",
		testEndpoint("/auth/login", 400, to{method: "POST", params: ms{"unique_id": uniqueID1}}))

	t.Run("User should be able to log in",
		testEndpoint("/auth/login", 200, to{method: "POST", params: login2, resCookies: &cookieslogout}))
	t.Run("Logged in user should be able to get own information",
		testEndpoint("/users/whoami", 200, to{cookies: cookieslogout, expectedUser: expectedUser{uniqueID: uniqueID2, role: ROLE_NONE}}))
	t.Run("User should be able to log out",
		testEndpoint("/auth/logout", 200, to{cookies: cookieslogout}))
	t.Run("Logged out user should not be able to get own information",
		testEndpoint("/users/whoami", 401, to{cookies: cookieslogout}))

	t.Run("Check APP State", checkAppState([]expectedUser{
		{uniqueID: uniqueID2, role: ROLE_NONE},
		{uniqueID: uniqueID3, role: ROLE_NONE},
		{uniqueID: uniqueID1, role: ROLE_ADMIN}}))

	userNIE := newUser("name", "asd3@example.com", uniqueID5, "12345678")
	t.Run("Registers with invalid id format should be rejected",
		testEndpoint("/auth/register", 401, to{method: "POST", params: userNIE}))

	t.Run("Admin can update site config",
		testEndpoint("/config/update", 200, to{method: "POST", cookies: cookies1, params: m{"id_formats": []string{ID_DNI, ID_NIE}}}))
	t.Run("Admin cannot remove id formats from site config",
		testEndpoint("/config/update", 500, to{method: "POST", cookies: cookies1, params: m{"id_formats": []string{ID_DNI}}}))

	t.Run("Registers with previously invalid id format should work",
		testEndpoint("/auth/register", 200, to{method: "POST", params: userNIE}))
	t.Run("NIE user should be able to log in",
		testEndpoint("/auth/login", 200, to{method: "POST", params: login5, resCookies: &cookies5}))

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
		testEndpoint("/users/files/own", 200, to{cookies: cookies2, expectedFiles: []expectedFile{}}))

	t.Run("User should be able to upload files",
		testEndpoint("/users/files/upload", 200, to{cookies: cookies2, file: expectedFile{description: "file", name: "testfile.txt"}}))
	t.Run("User should be able to upload files and get them renamed",
		testEndpoint("/users/files/upload", 200, to{cookies: cookies2, file: expectedFile{description: "file", name: "testfile.txt"}}))
	t.Run("User should be able to upload files and get them renamed",
		testEndpoint("/users/files/upload", 200, to{cookies: cookies2, file: expectedFile{description: "file", name: "testfile.txt"}}))
	t.Run("Another user should be able to upload files",
		testEndpoint("/users/files/upload", 200, to{cookies: cookies3, file: expectedFile{description: "file", name: "testfile.txt"}}))
	t.Run("User should not be able to upload files without description",
		testEndpoint("/users/files/upload", 400, to{cookies: cookies2, file: expectedFile{description: "", name: "testfile.txt"}}))

	t.Run("User should be able to get its files",
		testEndpoint("/users/files/own", 200, to{cookies: cookies2, expectedFiles: []expectedFile{
			{name: "testfile.txt", description: "file"}, {name: "testfile_1.txt", description: "file"}, {name: "testfile_2.txt", description: "file"}}}))
	t.Run("User uploaded files should appear in uploads folder", checkUploadsFolder([]string{"testfile.txt", "testfile_1.txt", "testfile_2.txt", "testfile_3.txt"}))

	t.Run("User should be able to delete its files",
		testEndpoint("/users/files/delete", 200, to{cookies: cookies2, query: "?id=2"}))
	t.Run("User should not be able to delete files with invalid parameters",
		testEndpoint("/users/files/delete", 400, to{cookies: cookies2, query: "?id=-1"}))

	t.Run("User deleted file should have disappeared",
		testEndpoint("/users/files/own", 200, to{cookies: cookies2, expectedFiles: []expectedFile{
			{name: "testfile.txt", description: "file"}, {name: "testfile_2.txt", description: "file"}}}))
	t.Run("User deleted file should have disappeared from uploads folder", checkUploadsFolder([]string{"testfile.txt", "testfile_2.txt", "testfile_3.txt"}))

	t.Run("User should be able to download its files",
		testEndpoint("/users/files/download", 200, to{cookies: cookies2, query: "?id=1", fileContent: "file content\n"}))

	t.Run("User should not be able to download deleted files",
		testEndpoint("/users/files/download", 401, to{cookies: cookies2, query: "?id=2"}))
	t.Run("User should not be able to download another user's files",
		testEndpoint("/users/files/download", 401, to{cookies: cookies2, query: "?id=4"}))
	t.Run("User should not be able to download files with invalid parameters",
		testEndpoint("/users/files/download", 400, to{cookies: cookies2, query: "?id=0"}))
	t.Run("User should not be able to delete another user's files",
		testEndpoint("/users/files/delete", 401, to{cookies: cookies2, query: "?id=4"}))

	t.Run("Admin should be able to download another user's files",
		testEndpoint("/users/files/download", 200, to{cookies: cookies1, query: "?id=4", fileContent: "file content\n"}))
	t.Run("Admin should be able to delete another user's files",
		testEndpoint("/users/files/delete", 200, to{cookies: cookies1, query: "?id=4"}))
	t.Run("User deleted file should have disappeared from uploads folder", checkUploadsFolder([]string{"testfile.txt", "testfile_2.txt"}))

	// User validation, including validation messages

	t.Run("Non-logged user should not get list of unvalidated users",
		testEndpoint("/users/unvalidated/get", 401, to{query: "?page=1&items_per_page=5"}))
	t.Run("Non-admin user should not get list of unvalidated users",
		testEndpoint("/users/unvalidated/get", 401, to{cookies: cookies2, query: "?page=1&items_per_page=5"}))
	t.Run("Admin user should get list of unvalidated users",
		testEndpoint("/users/unvalidated/get", 200, to{cookies: cookies1, query: "?page=1&items_per_page=5", expectedUsers: expectedUsersResponse{Total: 3, Users: []expectedUser{
			{uniqueID: uniqueID2}, {uniqueID: uniqueID3}, {uniqueID: uniqueID5}}}}))

	t.Run("Non-logged user should not be able to add messages",
		testEndpoint("/users/messages/add", 401, to{params: m{"user_id": 2, "content": "message content"}}))
	t.Run("Non-admin user should not be able to add messages",
		testEndpoint("/users/messages/add", 401, to{cookies: cookies2, params: m{"user_id": 2, "content": "message content"}}))
	t.Run("Admin user should be able to add messages",
		testEndpoint("/users/messages/add", 200, to{cookies: cookies1, params: m{"user_id": 2, "content": "message content user 2"}}))
	t.Run("Admin user should be able to add messages",
		testEndpoint("/users/messages/add", 200, to{cookies: cookies1, params: m{"user_id": 2, "content": "message content user 2"}}))
	t.Run("Admin user should be able to add messages",
		testEndpoint("/users/messages/add", 200, to{cookies: cookies1, params: m{"user_id": 3, "content": "message content user 3"}}))

	t.Run("List of unvalidated users should contain messages",
		testEndpoint("/users/unvalidated/get", 200, to{cookies: cookies1, query: "?page=1&items_per_page=5", expectedUsers: expectedUsersResponse{Total: 3, Users: []expectedUser{
			{uniqueID: uniqueID2, unsolvedMessages: []string{"message content user 2", "message content user 2"}},
			{uniqueID: uniqueID3, unsolvedMessages: []string{"message content user 3"}},
			{uniqueID: uniqueID5, unsolvedMessages: []string{}}}}}))

	t.Run("Non-logged user should not be able to get messages",
		testEndpoint("/users/messages/own", 401, to{}))
	t.Run("Logged user should get its own messages",
		testEndpoint("/users/messages/own", 200, to{cookies: cookies2, expectedUnsolvedMessages: []string{"message content user 2", "message content user 2"}}))

	t.Run("Non-logged user should not be able to solve messages",
		testEndpoint("/users/messages/solve", 401, to{query: "?id=1"}))
	t.Run("Logged user should be able to solve its messages",
		testEndpoint("/users/messages/solve", 200, to{cookies: cookies2, query: "?id=1"}))
	t.Run("Logged user should not be able to solve another user's messages",
		testEndpoint("/users/messages/solve", 401, to{cookies: cookies2, query: "?id=3"}))
	t.Run("Logged user should not be able to solve messages with invalid parameters",
		testEndpoint("/users/messages/solve", 400, to{cookies: cookies2, query: "?id=-123"}))
	t.Run("Admin user should be able to solve another user's messages",
		testEndpoint("/users/messages/solve", 200, to{cookies: cookies1, query: "?id=3"}))

	t.Run("Validated messages should not appear",
		testEndpoint("/users/messages/own", 200, to{cookies: cookies2, expectedUnsolvedMessages: []string{"message content user 2"}}))
	t.Run("Validated messages should not appear",
		testEndpoint("/users/messages/own", 200, to{cookies: cookies3, expectedUnsolvedMessages: []string{}}))

	t.Run("The only validated user should be admin",
		testEndpoint("/users/validated/get", 200, to{cookies: cookies1, query: "?page=1&items_per_page=5", expectedUsers: expectedUsersResponse{Total: 1, Users: []expectedUser{{uniqueID: uniqueID1}}}}))

	t.Run("Non-logged user should not be able to validate users",
		testEndpoint("/users/validate", 401, to{query: "?id=2"}))
	t.Run("Non-admin user should not be able to validate users",
		testEndpoint("/users/validate", 401, to{cookies: cookies2, query: "?id=2"}))
	t.Run("Admin user should be able to validate users with invalid paramters",
		testEndpoint("/users/validate", 400, to{cookies: cookies1, query: "?id=0"}))
	t.Run("Admin user should be able to validate users",
		testEndpoint("/users/validate", 200, to{cookies: cookies1, query: "?id=2"}))

	t.Run("Unexisting users cannot be validated",
		testEndpoint("/users/validate", 500, to{cookies: cookies1, query: "?id=999"}))
	t.Run("Validated users cannot be validated again",
		testEndpoint("/users/validate", 500, to{cookies: cookies1, query: "?id=2"}))
	t.Run("Admin users cannot be validated",
		testEndpoint("/users/validate", 500, to{cookies: cookies1, query: "?id=1"}))

	t.Run("Check APP State", checkAppState([]expectedUser{
		{uniqueID: uniqueID3, role: ROLE_NONE},
		{uniqueID: uniqueID2, role: ROLE_VALIDATED},
		{uniqueID: uniqueID5, role: ROLE_NONE},
		{uniqueID: uniqueID1, role: ROLE_ADMIN}}))

	t.Run("There should be two validated users",
		testEndpoint("/users/validated/get", 200, to{cookies: cookies1, query: "?page=1&items_per_page=5", expectedUsers: expectedUsersResponse{Total: 2, Users: []expectedUser{
			{uniqueID: uniqueID1}, {uniqueID: uniqueID2, unsolvedMessages: []string{"message content user 2"}}}}}))

	candidate1 := Candidate{Name: "candidate 1", Presentation: "candidate 1 presentation", Image: "candidate.jpg"}
	candidate2 := Candidate{Name: "candidate 2", Presentation: "candidate 2 presentation", Image: "candidate.jpg"}
	candidateWrong := Candidate{Name: "", Presentation: "candidate wrong presentation", Image: "candidate.jpg"}

	t.Run("Non-logged users should not be able to add candidates",
		testEndpoint("/candidates/add", 401, to{candidate: candidate1}))
	t.Run("Non-admin users should not be able to add candidates",
		testEndpoint("/candidates/add", 401, to{cookies: cookies2, candidate: candidate1}))
	t.Run("Admin users should not be able to add candidates with invalid parameters",
		testEndpoint("/candidates/add", 400, to{cookies: cookies1, candidate: candidateWrong}))
	t.Run("Admin users should be able to add candidates",
		testEndpoint("/candidates/add", 200, to{cookies: cookies1, candidate: candidate1}))
	t.Run("Admin users should be able to add candidates",
		testEndpoint("/candidates/add", 200, to{cookies: cookies1, candidate: candidate2}))

	candidate2.Image = "candidate_1.jpg"
	t.Run("Non-logged users should be able to get candidates",
		testEndpoint("/candidates/get", 200, to{expectedCandidates: []Candidate{candidate1, candidate2}}))
	t.Run("Logged users should be able to get candidates",
		testEndpoint("/candidates/get", 200, to{cookies: cookies2, expectedCandidates: []Candidate{candidate1, candidate2}}))
	t.Run("Candidate images should appear in uploads folder", checkUploadsFolder([]string{"testfile.txt", "testfile_2.txt", "candidate.jpg", "candidate_1.jpg"}))

	t.Run("Non-logged users should not be able to delete candidates",
		testEndpoint("/candidates/delete", 401, to{query: "?id=2"}))
	t.Run("Non-admin users should not be able to delete candidates",
		testEndpoint("/candidates/delete", 401, to{cookies: cookies2, query: "?id=2"}))
	t.Run("Admin users should be able to delete candidates",
		testEndpoint("/candidates/delete", 200, to{cookies: cookies1, query: "?id=2"}))

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
		testEndpoint("/elections/get", 200, to{cookies: cookies2, expectedElections: []Election{}}))
	t.Run("Admin user should be able to see unpublished elections",
		testEndpoint("/elections/get", 200, to{cookies: cookies1, expectedElections: []Election{election}}))

	t.Run("Non-logged user should not be able to publish election",
		testEndpoint("/elections/publish", 401, to{query: "?id=1"}))
	t.Run("Non-admin user should not be able to publish election",
		testEndpoint("/elections/publish", 401, to{cookies: cookies2, query: "?id=1"}))
	t.Run("Admin user should be able to publish election",
		testEndpoint("/elections/publish", 200, to{cookies: cookies1, query: "?id=1"}))

	election.Public = true
	t.Run("Non-logged user should be able to see elections",
		testEndpoint("/elections/get", 200, to{expectedElections: []Election{election}}))

	// more candidates for the election
	candidate2, candidate3, candidate4 := candidate1, candidate1, candidate1
	candidate2.Name = "candidate 2"
	candidate2.ID = 3
	t.Run("Admin users should be able to add candidates",
		testEndpoint("/candidates/add", 200, to{cookies: cookies1, candidate: candidate2}))
	candidate2.Image = "candidate_1.jpg"

	candidate3.Name = "candidate 3"
	candidate3.ID = 4
	t.Run("Admin users should be able to add candidates",
		testEndpoint("/candidates/add", 200, to{cookies: cookies1, candidate: candidate3}))
	candidate3.Image = "candidate_2.jpg"

	candidate4.Name = "candidate 4"
	candidate4.ID = 5
	t.Run("Admin users should be able to add candidates",
		testEndpoint("/candidates/add", 200, to{cookies: cookies1, candidate: candidate4}))
	candidate4.Image = "candidate_3.jpg"

	election.Candidates = append(election.Candidates, candidate2)
	election.Candidates = append(election.Candidates, candidate3)
	election.Candidates = append(election.Candidates, candidate4)

	t.Run("Admin user should not be able to vote before election start",
		testEndpoint("/elections/vote", 500, to{cookies: cookies1, params: m{"candidates": []int{1, 4}}}))
	timeTravel(90 * time.Minute)

	t.Run("Candidates should not be added after election starts",
		testEndpoint("/candidates/add", 401, to{cookies: cookies1, candidate: candidate1}))
	t.Run("Candidates should not be deleted after election starts",
		testEndpoint("/candidates/delete", 401, to{cookies: cookies1, query: "?id=1"}))

	var voteToken string
	t.Run("Admin user should be able to vote in time",
		testEndpoint("/elections/vote", 200, to{cookies: cookies1, params: m{"candidates": []int{1, 4}}, voteToken: &voteToken}))
	t.Run("Admin user should not be able to vote twice",
		testEndpoint("/elections/vote", 500, to{cookies: cookies1, params: m{"candidates": []int{1, 4}}}))
	t.Run("Admin user should be able to validate its vote",
		testEndpoint("/elections/vote/check", 200, to{params: m{"token": voteToken}, expectedCandidates: []Candidate{candidate1, candidate3}}))

	t.Run("Unvalidated user should not be able to vote",
		testEndpoint("/elections/vote", 401, to{cookies: cookies3, params: m{"candidates": []int{1, 4}}}))
	t.Run("Validated user should not be able to vote more than the maximum allowed candidates",
		testEndpoint("/elections/vote", 500, to{cookies: cookies2, params: m{"candidates": []int{1, 4, 5, 6}}}))
	t.Run("Validated user should not be able to vote less than the minimum allowed candidates",
		testEndpoint("/elections/vote", 500, to{cookies: cookies2, params: m{"candidates": []int{1}}}))
	t.Run("Validated user should not be able to vote unexisting candidates",
		testEndpoint("/elections/vote", 500, to{cookies: cookies2, params: m{"candidates": []int{-1, -2}}}))

	t.Run("Validated user should be able to vote just once", testVoteOnce(to{cookies: cookies2, params: m{"candidates": []int{3, 4}}}))

	t.Run("Admin user should be able to validate users",
		testEndpoint("/users/validate", 200, to{cookies: cookies1, query: "?id=3"}))
	t.Run("Admin user should be able to validate users",
		testEndpoint("/users/validate", 200, to{cookies: cookies1, query: "?id=4"})) // user with ID 4 has uniqueID5
	t.Run("Two users should be able to vote concurrently", testVoteConcurrent(
		to{cookies: cookies3, params: m{"candidates": []int{4, 5, 3}}},
		to{cookies: cookies5, params: m{"candidates": []int{5, 1, 3}}}))

	// see that elections can have its votes counted
	t.Run("The election should not have its votes counted yet",
		testEndpoint("/elections/get", 200, to{cookies: cookies1, expectedElections: []Election{election}}))
	checkElectionsCount()
	t.Run("The election should not have its votes counted yet",
		testEndpoint("/elections/get", 200, to{cookies: cookies1, expectedElections: []Election{election}}))

	timeTravel(60 * time.Minute) // election ended
	checkElectionsCount()
	election.Counted = true
	election.Candidates[0].Points = 7
	election.Candidates[1].Points = 8
	election.Candidates[2].Points = 10
	election.Candidates[3].Points = 7
	t.Run("The election should have its votes counted",
		testEndpoint("/elections/get", 200, to{cookies: cookies1, expectedElections: []Election{election}}))
	checkElectionsCount()
	t.Run("The election should have its votes counted",
		testEndpoint("/elections/get", 200, to{cookies: cookies1, expectedElections: []Election{election}}))
}

func testVoteOnce(options testOptions) func(*testing.T) {
	return func(t *testing.T) {
		ch := make(chan *httptest.ResponseRecorder)
		go func() { ch <- testEndpointAux(t, "/elections/vote", options, -1) }()
		go func() { ch <- testEndpointAux(t, "/elections/vote", options, -2) }()
		rr1 := <-ch
		rr2 := <-ch
		if !(rr1.Code == 200 && rr2.Code == 500) && !(rr1.Code == 500 && rr2.Code == 200) {
			t.Errorf("Same user trying to vote twice should result in a success and a failure, instead got (%d, %d)", rr1.Code, rr2.Code)
		}
	}
}

func testVoteConcurrent(options1, options2 testOptions) func(*testing.T) {
	return func(t *testing.T) {
		ch := make(chan *httptest.ResponseRecorder)
		go func() { ch <- testEndpointAux(t, "/elections/vote", options1, -3) }()
		go func() { ch <- testEndpointAux(t, "/elections/vote", options2, -4) }()
		rr1 := <-ch
		rr2 := <-ch
		if rr1.Code != 200 || rr2.Code != 200 {
			t.Errorf("Two users trying to vote at the same time should work, instead got (%d, %d)", rr1.Code, rr2.Code)
		}
	}
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
		rr := testEndpointAux(t, path, options, i)
		if rr.Code != expectedCode {
			t.Errorf("[%d] Expected code %v testing endpoint %q, but got %v.", i, expectedCode, path, rr.Code)
		}

		if options.expectedUsers.Users != nil {
			var response getUsersResponse
			if err := json.Unmarshal([]byte(rr.Body.String()), &response); err != nil {
				t.Errorf("Could not unmarshal expected users response: %s", err)
			} else {
				compareUsers(t, options.expectedUsers.Users, response.Users)
				if options.expectedUsers.Total != response.Total {
					t.Errorf("Expected %d total users but got %d.", options.expectedUsers.Total, response.Total)
				}
			}
		}

		if options.expectedUser.uniqueID != "" {
			var response User
			if err := json.Unmarshal([]byte(rr.Body.String()), &response); err != nil {
				t.Errorf("Could not unmarshal expected user response: %s", err)
			} else {
				compareUsers(t, []expectedUser{options.expectedUser}, []User{response})
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

		if options.voteToken != nil {
			*options.voteToken = strings.Trim(rr.Body.String(), "\"")
		}

		if options.fileContent != "" && options.fileContent != rr.Body.String() {
			t.Errorf("Wrong file contents. Expected %q but found %q.", options.fileContent, rr.Body.String())
		}
	}
}

func testEndpointAux(t *testing.T, path string, options testOptions, i int) *httptest.ResponseRecorder {
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
	} else if options.postParams != nil {
		b := &bytes.Buffer{}
		writer := multipart.NewWriter(b)
		for key, val := range options.postParams {
			if err = writer.WriteField(key, val); err != nil {
				t.Fatalf("[%d] Could not write form field for endpoint %q. Error: %s", i, path, err)
			}
		}
		if err := writer.Close(); err != nil {
			t.Fatalf("[%d] Could not close writer for endpoint %q. Error: %s", i, path, err)
		}
		body = b
		contentType = writer.FormDataContentType()
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

	if options.cookies != nil {
		for _, cookie := range options.cookies {
			req.AddCookie(cookie)
		}
	}
	req.Header.Set("Content-Type", contentType)

	rr := httptest.NewRecorder()
	h, ok := appHandlers[path]
	if !ok {
		t.Fatalf("Unknown handler for path %q", path)
	}

	handler := http.HandlerFunc(h)
	handler.ServeHTTP(rr, req)
	if options.resCookies != nil {
		for _, cookie := range rr.Result().Cookies() {
			*options.resCookies = append(*options.resCookies, cookie)
		}
	}

	return rr
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
		for k := range expected {
			sort.Slice(expected[k].Candidates, func(i, j int) bool { return expected[k].Candidates[i].ID < expected[k].Candidates[j].ID })
			sort.Slice(got[k].Candidates, func(i, j int) bool { return got[k].Candidates[i].ID < got[k].Candidates[j].ID })
		}
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
