package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"testing"
)

const baseURL = "http://localhost:8080"

// ── HTTP helpers ─────────────────────────────────────────────────────────────

func post(t *testing.T, path string, body any, token string) (int, map[string]any) {
	t.Helper()
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, baseURL+path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	return readResp(t, resp)
}

func get(t *testing.T, path string, token string) (int, map[string]any) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, baseURL+path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	return readResp(t, resp)
}

func readResp(t *testing.T, resp *http.Response) (int, map[string]any) {
	t.Helper()
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var result map[string]any
	_ = json.Unmarshal(raw, &result)
	return resp.StatusCode, result
}

func getArray(t *testing.T, path string, token string) (int, []any) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, baseURL+path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var result []any
	_ = json.Unmarshal(raw, &result)
	return resp.StatusCode, result
}

func postArray(t *testing.T, path string, body any, token string) (int, []any) {
	t.Helper()
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, baseURL+path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var result []any
	_ = json.Unmarshal(raw, &result)
	return resp.StatusCode, result
}

func put(t *testing.T, path string, body any, token string) (int, map[string]any) {
	t.Helper()
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPut, baseURL+path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT %s: %v", path, err)
	}
	return readResp(t, resp)
}

func patch(t *testing.T, path string, body any, token string) (int, map[string]any) {
	t.Helper()
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPatch, baseURL+path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PATCH %s: %v", path, err)
	}
	return readResp(t, resp)
}

func del(t *testing.T, path string, token string) (int, map[string]any) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodDelete, baseURL+path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE %s: %v", path, err)
	}
	return readResp(t, resp)
}

// login registers (or logs in) a phone and returns the token pair.
// Code is always "1234" in this environment.
func login(t *testing.T, phone string) (accessToken, refreshToken string, isNewUser bool) {
	t.Helper()
	status, _ := post(t, "/api/v1/auth/send_code", map[string]any{"Phone": phone}, "")
	if status != http.StatusOK {
		t.Fatalf("send_code: expected 200, got %d", status)
	}

	status, body := post(t, "/api/v1/auth/verify_code", map[string]any{"phone": phone, "code": "1234"}, "")
	if status != http.StatusOK {
		t.Fatalf("verify_code: expected 200, got %d (body: %v)", status, body)
	}

	accessToken, _ = body["access_token"].(string)
	refreshToken, _ = body["refresh_token"].(string)
	isNewUser, _ = body["is_new_user"].(bool)

	if accessToken == "" || refreshToken == "" {
		t.Fatalf("verify_code: missing tokens in response: %v", body)
	}
	return
}

// requireAI skips tests that need a real AI provider when no GigaChat key is
// available (e.g. pull requests from forks, where CI secrets are not exposed).
func requireAI(t *testing.T) {
	t.Helper()
	if os.Getenv("GIGACHAT_AUTHORIZATION_KEY") == "" {
		t.Skip("GIGACHAT_AUTHORIZATION_KEY is not set: skipping test that needs a real AI provider")
	}
}

// newPhone generates a unique phone per call using a random suffix so tests
// don't collide across runs against a persistent database.
func newPhone() string {
	return fmt.Sprintf("+7%010d", rand.Int63n(9_000_000_000)+1_000_000_000)
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestSendCode_InvalidPhone(t *testing.T) {
	status, body := post(t, "/api/v1/auth/send_code", map[string]any{"Phone": "not-a-phone"}, "")
	// send_code does not validate phone format itself — it just stores and sends.
	// This test documents the current behaviour.
	t.Logf("send_code invalid phone → %d %v", status, body)
}

func TestVerifyCode_WrongCode(t *testing.T) {
	phone := newPhone()
	post(t, "/api/v1/auth/send_code", map[string]any{"Phone": phone}, "")

	status, body := post(t, "/api/v1/auth/verify_code", map[string]any{"phone": phone, "code": "0000"}, "")
	if status != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d (body: %v)", status, body)
	}
}

func TestVerifyCode_InvalidFormat(t *testing.T) {
	// phone not E.164, code not 4 digits
	status, body := post(t, "/api/v1/auth/verify_code", map[string]any{"phone": "abc", "code": "12"}, "")
	if status != http.StatusBadRequest {
		t.Errorf("expected 400, got %d (body: %v)", status, body)
	}
}

func TestLogin_NewUser(t *testing.T) {
	_, _, isNewUser := login(t, newPhone())
	if !isNewUser {
		t.Error("expected is_new_user=true for a fresh phone number")
	}
}

func TestLogin_ExistingUser(t *testing.T) {
	phone := newPhone()
	login(t, phone) // first login — creates user

	_, _, isNewUser := login(t, phone) // second login
	if isNewUser {
		t.Error("expected is_new_user=false on second login")
	}
}

func TestRefreshToken(t *testing.T) {
	access, refresh, _ := login(t, newPhone())

	status, body := post(t, "/api/v1/auth/refresh_token", map[string]any{
		"access_token":  access,
		"refresh_token": refresh,
	}, "")
	if status != http.StatusOK {
		t.Fatalf("refresh_token: expected 200, got %d (body: %v)", status, body)
	}

	newAccess, _ := body["access_token"].(string)
	newRefresh, _ := body["refresh_token"].(string)
	if newAccess == "" || newRefresh == "" {
		t.Error("refresh_token: missing tokens in response")
	}
}

func TestRefreshToken_MismatchedPair(t *testing.T) {
	access1, _, _ := login(t, newPhone())
	_, refresh2, _ := login(t, newPhone())

	status, body := post(t, "/api/v1/auth/refresh_token", map[string]any{
		"access_token":  access1,
		"refresh_token": refresh2,
	}, "")
	if status != http.StatusUnauthorized {
		t.Errorf("expected 401 for mismatched token pair, got %d (body: %v)", status, body)
	}
}

func TestGetProfile_Unauthorized(t *testing.T) {
	status, _ := get(t, "/api/v1/profile/", "")
	if status != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", status)
	}
}

func TestGetProfile_NewUser(t *testing.T) {
	access, _, _ := login(t, newPhone())

	status, body := get(t, "/api/v1/profile/", access)
	if status != http.StatusOK {
		t.Fatalf("get profile: expected 200, got %d (body: %v)", status, body)
	}

	if body["UUID"] == nil {
		t.Error("profile missing UUID")
	}
	if body["Phone"] == nil {
		t.Error("profile missing Phone")
	}
	teeth, _ := body["Teeth"].([]any)
	if len(teeth) != 32 {
		t.Errorf("expected 32 teeth for new user, got %d", len(teeth))
	}
	for i, raw := range teeth {
		tooth := raw.(map[string]any)
		if tooth["ToothState"] != "Healthy" {
			t.Errorf("tooth[%d]: expected Healthy, got %v", i, tooth["ToothState"])
		}
		if tooth["ToothType"] != "Molar" {
			t.Errorf("tooth[%d]: expected Molar, got %v", i, tooth["ToothType"])
		}
	}
}

func TestUpdateProfile_Name(t *testing.T) {
	access, _, _ := login(t, newPhone())

	status, body := patch(t, "/api/v1/profile/", map[string]any{"name": "John"}, access)
	if status != http.StatusOK {
		t.Fatalf("update name: expected 200, got %d (body: %v)", status, body)
	}

	status, profile := get(t, "/api/v1/profile/", access)
	if status != http.StatusOK {
		t.Fatalf("get profile: expected 200, got %d", status)
	}
	if profile["Name"] != "John" {
		t.Errorf("name: expected John, got %v", profile["Name"])
	}
}

func TestUpdateProfile_AgeAndGender(t *testing.T) {
	access, _, _ := login(t, newPhone())

	age := 28
	status, body := patch(t, "/api/v1/profile/", map[string]any{
		"age":    age,
		"gender": "Female",
	}, access)
	if status != http.StatusOK {
		t.Fatalf("update profile: expected 200, got %d (body: %v)", status, body)
	}

	status, profile := get(t, "/api/v1/profile/", access)
	if status != http.StatusOK {
		t.Fatalf("get profile after update: expected 200, got %d", status)
	}

	gotAge, _ := profile["Age"].(float64)
	if int(gotAge) != age {
		t.Errorf("age: expected %d, got %v", age, profile["Age"])
	}
	if profile["Gender"] != "Female" {
		t.Errorf("gender: expected Female, got %v", profile["Gender"])
	}
}

// teethBody builds the PATCH /profile/teeth request body from a list of states.
func teethBody(states []string) map[string]any {
	items := make([]map[string]any, len(states))
	for i, st := range states {
		items[i] = map[string]any{"tooth_state": st}
	}
	return map[string]any{"teeth": items}
}

func TestUpdateProfile_TeethState(t *testing.T) {
	access, _, _ := login(t, newPhone())

	// Tooth 0 is Problematic, tooth 5 is Missing, the rest are Healthy
	states := make([]string, 32)
	for i := range states {
		states[i] = "Healthy"
	}
	states[0] = "Problematic"
	states[5] = "Missing"

	status, body := patch(t, "/api/v1/profile/teeth", teethBody(states), access)
	if status != http.StatusOK {
		t.Fatalf("update teeth: expected 200, got %d (body: %v)", status, body)
	}

	status, profile := get(t, "/api/v1/profile/", access)
	if status != http.StatusOK {
		t.Fatalf("get profile: expected 200, got %d", status)
	}

	teeth, _ := profile["Teeth"].([]any)
	if len(teeth) != 32 {
		t.Fatalf("expected 32 teeth, got %d", len(teeth))
	}

	// teeth come back sorted by tooth_index from DB query
	toothByIndex := map[int]map[string]any{}
	for _, raw := range teeth {
		tooth := raw.(map[string]any)
		idx := int(tooth["ToothIndex"].(float64))
		toothByIndex[idx] = tooth
	}

	for i := 0; i < 32; i++ {
		expected := states[i]
		got := toothByIndex[i]["ToothState"]
		if got != expected {
			t.Errorf("tooth[%d]: expected %s, got %v", i, expected, got)
		}
	}
}

func TestUpdateProfile_TimesBrush(t *testing.T) {
	access, _, _ := login(t, newPhone())

	status, body := put(t, "/api/v1/profile/brush-schedule", map[string]any{
		"times_brush": []string{"08:00", "21:30"},
	}, access)
	if status != http.StatusOK {
		t.Fatalf("update brush times: expected 200, got %d (body: %v)", status, body)
	}

	status, profile := get(t, "/api/v1/profile/", access)
	if status != http.StatusOK {
		t.Fatalf("get profile: expected 200, got %d", status)
	}

	times, _ := profile["BrushTimes"].([]any)
	if len(times) != 2 {
		t.Fatalf("expected 2 brush times, got %d", len(times))
	}
}

func TestUpdateProfile_TimesBrush_Replace(t *testing.T) {
	access, _, _ := login(t, newPhone())

	// Set initial times
	put(t, "/api/v1/profile/brush-schedule", map[string]any{"times_brush": []string{"07:00", "12:00", "22:00"}}, access)

	// Replace with different times
	status, body := put(t, "/api/v1/profile/brush-schedule", map[string]any{
		"times_brush": []string{"09:00"},
	}, access)
	if status != http.StatusOK {
		t.Fatalf("replace brush times: expected 200, got %d (body: %v)", status, body)
	}

	status, profile := get(t, "/api/v1/profile/", access)
	if status != http.StatusOK {
		t.Fatalf("get profile: expected 200, got %d", status)
	}

	times, _ := profile["BrushTimes"].([]any)
	if len(times) != 1 {
		t.Errorf("expected 1 brush time after replace, got %d", len(times))
	}
}

func TestUpdateProfile_InvalidTeethCount(t *testing.T) {
	access, _, _ := login(t, newPhone())

	// Only 5 teeth instead of 32
	status, body := patch(t, "/api/v1/profile/teeth",
		teethBody([]string{"Healthy", "Healthy", "Healthy", "Healthy", "Healthy"}), access)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for wrong teeth count, got %d (body: %v)", status, body)
	}
}

func TestUpdateProfile_InvalidToothState(t *testing.T) {
	access, _, _ := login(t, newPhone())

	states := make([]string, 32)
	for i := range states {
		states[i] = "Healthy"
	}
	states[3] = "Unhealthy" // not a valid state: Healthy, Problematic or Missing

	status, body := patch(t, "/api/v1/profile/teeth", teethBody(states), access)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid tooth state, got %d (body: %v)", status, body)
	}
}

func TestUpdateProfile_ToothNote(t *testing.T) {
	access, _, _ := login(t, newPhone())

	items := make([]map[string]any, 32)
	for i := range items {
		items[i] = map[string]any{"tooth_state": "Healthy"}
	}
	items[7] = map[string]any{"tooth_state": "Problematic", "tooth_note": "sensitive to cold"}

	status, body := patch(t, "/api/v1/profile/teeth", map[string]any{"teeth": items}, access)
	if status != http.StatusOK {
		t.Fatalf("update teeth: expected 200, got %d (body: %v)", status, body)
	}

	status, profile := get(t, "/api/v1/profile/", access)
	if status != http.StatusOK {
		t.Fatalf("get profile: expected 200, got %d", status)
	}
	teeth, _ := profile["Teeth"].([]any)
	for _, raw := range teeth {
		tooth := raw.(map[string]any)
		if int(tooth["ToothIndex"].(float64)) == 7 {
			if tooth["ToothNote"] != "sensitive to cold" {
				t.Errorf("tooth 7: expected note %q, got %v", "sensitive to cold", tooth["ToothNote"])
			}
			return
		}
	}
	t.Error("tooth 7 not found in profile")
}

func TestUpdateProfile_InvalidGender(t *testing.T) {
	access, _, _ := login(t, newPhone())

	status, body := patch(t, "/api/v1/profile/", map[string]any{"gender": "Unknown"}, access)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid gender, got %d (body: %v)", status, body)
	}
}

func TestUpdateProfile_InvalidBrushTimeFormat(t *testing.T) {
	access, _, _ := login(t, newPhone())

	status, body := put(t, "/api/v1/profile/brush-schedule", map[string]any{"times_brush": []string{"9am", "not-a-time"}}, access)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid time format, got %d (body: %v)", status, body)
	}
}

// ── Doctor feature tests ──────────────────────────────────────────────────────

const fakePatientUUID = "00000000-0000-0000-0000-000000000001"
const fakeNoteUUID = "00000000-0000-0000-0000-000000000002"

// doctorPaths lists the doctor endpoints that must enforce authentication and role.
var doctorGetPaths = []string{
	"/api/v1/doctor/patients/" + fakePatientUUID,
	"/api/v1/doctor/patients/" + fakePatientUUID + "/teeth",
	"/api/v1/doctor/patients/" + fakePatientUUID + "/teeth/notes",
	"/api/v1/doctor/patients/" + fakePatientUUID + "/recommendations",
}

func TestDoctorGetEndpoints_Unauthorized(t *testing.T) {
	for _, path := range doctorGetPaths {
		path := path
		t.Run(path, func(t *testing.T) {
			status, _ := get(t, path, "")
			if status != http.StatusUnauthorized {
				t.Errorf("expected 401, got %d", status)
			}
		})
	}
}

func TestDoctorGetEndpoints_ForbiddenForRegularUser(t *testing.T) {
	access, _, _ := login(t, newPhone())
	for _, path := range doctorGetPaths {
		path := path
		t.Run(path, func(t *testing.T) {
			status, _ := get(t, path, access)
			if status != http.StatusForbidden {
				t.Errorf("expected 403, got %d", status)
			}
		})
	}
}

func TestDoctorPostEndpoints_Unauthorized(t *testing.T) {
	cases := []struct{ path string }{
		{"/api/v1/doctor/patients/" + fakePatientUUID + "/teeth/notes"},
		{"/api/v1/doctor/patients/" + fakePatientUUID + "/recommendations"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.path, func(t *testing.T) {
			status, _ := post(t, tc.path, map[string]any{"note": "x"}, "")
			if status != http.StatusUnauthorized {
				t.Errorf("expected 401, got %d", status)
			}
		})
	}
}

func TestDoctorPostEndpoints_ForbiddenForRegularUser(t *testing.T) {
	access, _, _ := login(t, newPhone())
	cases := []struct{ path string }{
		{"/api/v1/doctor/patients/" + fakePatientUUID + "/teeth/notes"},
		{"/api/v1/doctor/patients/" + fakePatientUUID + "/recommendations"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.path, func(t *testing.T) {
			status, _ := post(t, tc.path, map[string]any{"note": "x"}, access)
			if status != http.StatusForbidden {
				t.Errorf("expected 403, got %d", status)
			}
		})
	}
}

func TestDoctorPutDelete_Unauthorized(t *testing.T) {
	notePath := "/api/v1/doctor/patients/" + fakePatientUUID + "/teeth/notes/" + fakeNoteUUID

	status, _ := put(t, notePath, map[string]any{"note": "x"}, "")
	if status != http.StatusUnauthorized {
		t.Errorf("PUT without token: expected 401, got %d", status)
	}

	status, _ = del(t, notePath, "")
	if status != http.StatusUnauthorized {
		t.Errorf("DELETE without token: expected 401, got %d", status)
	}
}

func TestDoctorPutDelete_ForbiddenForRegularUser(t *testing.T) {
	access, _, _ := login(t, newPhone())
	notePath := "/api/v1/doctor/patients/" + fakePatientUUID + "/teeth/notes/" + fakeNoteUUID

	status, _ := put(t, notePath, map[string]any{"note": "x"}, access)
	if status != http.StatusForbidden {
		t.Errorf("PUT regular user: expected 403, got %d", status)
	}

	status, _ = del(t, notePath, access)
	if status != http.StatusForbidden {
		t.Errorf("DELETE regular user: expected 403, got %d", status)
	}
}

// ── Chat feature tests ────────────────────────────────────────────────────────

func TestChat_Unauthorized(t *testing.T) {
	status, _ := post(t, "/api/v1/ai-chat/message", map[string]any{"message": "hello"}, "")
	if status != http.StatusUnauthorized {
		t.Errorf("POST /ai-chat/message without token: expected 401, got %d", status)
	}

	status, _ = get(t, "/api/v1/ai-chat/history", "")
	if status != http.StatusUnauthorized {
		t.Errorf("GET /ai-chat/history without token: expected 401, got %d", status)
	}
}

func TestChat_SendMessage(t *testing.T) {
	requireAI(t)
	access, _, _ := login(t, newPhone())

	status, msgs := postArray(t, "/api/v1/ai-chat/message", map[string]any{"message": "how often should I floss?"}, access)
	if status != http.StatusOK {
		t.Fatalf("send message: expected 200, got %d (body: %v)", status, msgs)
	}

	// the response contains the saved user message followed by the assistant reply
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages (user + assistant), got %d", len(msgs))
	}
	user, _ := msgs[0].(map[string]any)
	reply, _ := msgs[1].(map[string]any)
	if user["Role"] != "user" {
		t.Errorf("expected first Role=user, got %v", user["Role"])
	}
	if reply["Role"] != "assistant" {
		t.Errorf("expected second Role=assistant, got %v", reply["Role"])
	}
	if reply["Content"] == nil || reply["Content"] == "" {
		t.Error("expected non-empty Content in assistant reply")
	}
}

func TestChat_HistoryGrowsAfterMessages(t *testing.T) {
	requireAI(t)
	access, _, _ := login(t, newPhone())

	// empty history for new user
	status, msgs := getArray(t, "/api/v1/ai-chat/history", access)
	if status != http.StatusOK {
		t.Fatalf("get history: expected 200, got %d", status)
	}
	if len(msgs) != 0 {
		t.Errorf("expected empty history for new user, got %d entries", len(msgs))
	}

	// send two messages → 4 entries (user + assistant each)
	post(t, "/api/v1/ai-chat/message", map[string]any{"message": "what is a cavity?"}, access)
	post(t, "/api/v1/ai-chat/message", map[string]any{"message": "how to prevent gum disease?"}, access)

	status, msgs = getArray(t, "/api/v1/ai-chat/history", access)
	if status != http.StatusOK {
		t.Fatalf("get history after messages: expected 200, got %d", status)
	}
	if len(msgs) != 4 {
		t.Errorf("expected 4 history entries after 2 messages, got %d", len(msgs))
	}
}

// ── Messaging feature tests ───────────────────────────────────────────────────

func TestMessaging_PatientConversations_Unauthorized(t *testing.T) {
	status, _ := get(t, "/api/v1/profile/conversations", "")
	if status != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", status)
	}
}

func TestMessaging_DoctorConversations_Unauthorized(t *testing.T) {
	status, _ := get(t, "/api/v1/doctor/conversations", "")
	if status != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", status)
	}
}

func TestMessaging_DoctorConversations_ForbiddenForPatient(t *testing.T) {
	access, _, _ := login(t, newPhone())
	status, _ := get(t, "/api/v1/doctor/conversations", access)
	if status != http.StatusForbidden {
		t.Errorf("expected 403 for patient, got %d", status)
	}
}

func TestMessaging_PatientCanListEmptyConversations(t *testing.T) {
	access, _, _ := login(t, newPhone())
	status, msgs := getArray(t, "/api/v1/profile/conversations", access)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if len(msgs) != 0 {
		t.Errorf("expected empty list for new patient, got %d", len(msgs))
	}
}

func TestMessaging_SendMessage_NotParticipant(t *testing.T) {
	access, _, _ := login(t, newPhone())
	fakeConvID := "00000000-0000-0000-0000-000000000099"

	status, _ := post(t, "/api/v1/profile/conversations/"+fakeConvID+"/messages",
		map[string]any{"content": "hi"}, access)
	// 404 — conversation doesn't exist
	if status != http.StatusNotFound {
		t.Errorf("expected 404 for unknown conversation, got %d", status)
	}
}

func TestMessaging_SendMessage_InvalidConversationUUID(t *testing.T) {
	access, _, _ := login(t, newPhone())
	status, _ := post(t, "/api/v1/profile/conversations/not-a-uuid/messages",
		map[string]any{"content": "hi"}, access)
	if status != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", status)
	}
}

func TestChat_InvalidMessage(t *testing.T) {
	access, _, _ := login(t, newPhone())

	status, body := post(t, "/api/v1/ai-chat/message", map[string]any{"message": ""}, access)
	if status != http.StatusBadRequest {
		t.Errorf("empty message: expected 400, got %d (body: %v)", status, body)
	}
}
