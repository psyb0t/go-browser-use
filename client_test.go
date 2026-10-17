package browseruse_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	browseruse "github.com/psyb0t/go-browser-use"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAPIKey       = "EXAMPLE-DO-NOT-USE"
	testRunID        = "00000000-0000-4000-8000-000000000001"
	testSessionID    = "00000000-0000-4000-8000-000000000002"
	testWorkspaceID  = "00000000-0000-4000-8000-000000000003"
	apiPrefix        = "/api/v4"
	browserUseHeader = "X-Browser-Use-API-Key"
	contentTypeJSON  = "application/json"
)

func newTestClient(
	t *testing.T,
	handler http.HandlerFunc,
) *browseruse.ClientWithResponses {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	auth := browseruse.WithRequestEditorFn(
		func(_ context.Context, req *http.Request) error {
			req.Header.Set(browserUseHeader, testAPIKey)

			return nil
		},
	)

	client, err := browseruse.NewClientWithResponses(srv.URL+apiPrefix, auth)
	require.NoError(t, err)

	return client
}

func writeJSON(t *testing.T, w http.ResponseWriter, body any) {
	t.Helper()

	w.Header().Set("Content-Type", contentTypeJSON)
	assert.NoError(t, json.NewEncoder(w).Encode(body))
}

func TestCreateRun(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		cost string
	}{
		{name: "number cost", cost: `1.23456789`},
		{name: "decimal string cost", cost: `"1.23456789"`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := newTestClient(t, func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, apiPrefix+"/runs", r.URL.Path)
				assert.Equal(t, testAPIKey, r.Header.Get(browserUseHeader))
				assert.Equal(t, contentTypeJSON, r.Header.Get("Content-Type"))

				var body map[string]json.RawMessage

				assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				assert.JSONEq(t, `"Read the page title"`, string(body["task"]))
				assert.Equal(t, tc.cost, string(body["maxCostUsd"]))
				assert.NotContains(t, body, "sessionId")

				writeJSON(t, w, browseruse.RunCreateResponse{
					Id:          uuid.MustParse(testRunID),
					SessionId:   uuid.MustParse(testSessionID),
					WorkspaceId: uuid.MustParse(testWorkspaceID),
					Status:      browseruse.RunCreateResponseStatusQueued,
					Model:       "example-model",
					EventsUrl:   apiPrefix + "/runs/" + testRunID + "/events",
				})
			})

			var cost browseruse.RunCreateRequest_MaxCostUsd

			require.NoError(t, json.Unmarshal([]byte(tc.cost), &cost))

			res, err := client.CreateRunRunsPostWithResponse(
				t.Context(),
				browseruse.RunCreateRequest{
					Task:       "Read the page title",
					MaxCostUsd: &cost,
				},
			)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, res.StatusCode())
			require.NotNil(t, res.JSON200)
			assert.Equal(t, testRunID, res.JSON200.Id.String())
			assert.Equal(t, testSessionID, res.JSON200.SessionId.String())
			assert.Equal(t, browseruse.RunCreateResponseStatusQueued,
				res.JSON200.Status)
		})
	}
}

func TestListRunsPagination(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, apiPrefix+"/runs", r.URL.Path)
		assert.Equal(t, testAPIKey, r.Header.Get(browserUseHeader))
		assert.Equal(t, "2", r.URL.Query().Get("limit"))

		response := browseruse.RunListResponse{Runs: []browseruse.RunSummary{}}
		hasMore := false
		cursor := "next-page"

		switch r.URL.Query().Get("cursor") {
		case "":
			hasMore = true
			response.NextCursor = &cursor
			response.Runs = []browseruse.RunSummary{
				{Id: uuid.MustParse(testRunID), TotalCostUsd: "0.012345"},
				{Id: uuid.MustParse(testSessionID), TotalCostUsd: "0.000001"},
			}
		case cursor:
			response.Runs = []browseruse.RunSummary{
				{Id: uuid.MustParse(testWorkspaceID), TotalCostUsd: "0.000002"},
			}
		case "empty":
		default:
			t.Errorf("unexpected cursor: %q", r.URL.Query().Get("cursor"))
		}

		response.HasMore = &hasMore
		writeJSON(t, w, response)
	})

	limit := 2
	params := &browseruse.ListRunsRunsGetParams{Limit: &limit}

	var ids []string

	for range 2 {
		res, err := client.ListRunsRunsGetWithResponse(t.Context(), params)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode())
		require.NotNil(t, res.JSON200)
		require.NotNil(t, res.JSON200.HasMore)

		for _, run := range res.JSON200.Runs {
			ids = append(ids, run.Id.String())
			assert.Nil(t, run.Title)
			assert.Nil(t, run.WorkspaceId)
		}

		if !*res.JSON200.HasMore {
			assert.Nil(t, res.JSON200.NextCursor)

			break
		}

		require.Len(t, res.JSON200.Runs, limit)
		require.NotNil(t, res.JSON200.NextCursor)
		assert.Equal(t, "0.012345", res.JSON200.Runs[0].TotalCostUsd)
		params.Cursor = res.JSON200.NextCursor
	}

	assert.Equal(t, []string{testRunID, testSessionID, testWorkspaceID}, ids)

	emptyCursor := "empty"
	params.Cursor = &emptyCursor
	res, err := client.ListRunsRunsGetWithResponse(t.Context(), params)
	require.NoError(t, err)
	require.NotNil(t, res.JSON200)
	assert.Empty(t, res.JSON200.Runs)
}

func TestRunErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		status int
		body   string
	}{
		{
			name: "forbidden", status: http.StatusForbidden,
			body: `{"detail":"Forbidden"}`,
		},
		{
			name: "validation", status: http.StatusUnprocessableEntity,
			body: `{"detail":[{
				"loc":["body","task"],"msg":"Required","type":"missing"
			}]}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := newTestClient(t, func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				assert.Equal(t, testAPIKey, r.Header.Get(browserUseHeader))
				w.Header().Set("Content-Type", contentTypeJSON)
				w.WriteHeader(tc.status)
				_, err := w.Write([]byte(tc.body))
				assert.NoError(t, err)
			})

			res, err := client.CreateRunRunsPostWithResponse(
				t.Context(), browseruse.RunCreateRequest{Task: "Example task"},
			)
			require.NoError(t, err)
			assert.Equal(t, tc.status, res.StatusCode())
			assert.Nil(t, res.JSON200)
			assert.JSONEq(t, tc.body, string(res.Body))

			if tc.status == http.StatusUnprocessableEntity {
				require.NotNil(t, res.JSON422)
				require.NotNil(t, res.JSON422.Detail)
				assert.Len(t, *res.JSON422.Detail, 1)
			}
		})
	}
}

func TestSessionExplicitNull(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, apiPrefix+"/sessions/"+testSessionID, r.URL.Path)
		assert.Equal(t, testAPIKey, r.Header.Get(browserUseHeader))

		var body map[string]json.RawMessage

		assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "null", string(body["title"]))
		writeJSON(t, w, browseruse.SessionInfo{
			SessionId:   uuid.MustParse(testSessionID),
			LatestRunId: uuid.MustParse(testRunID),
			Task:        "Example task",
			Status:      browseruse.SessionInfoStatusCompleted,
		})
	})

	res, err := client.UpdateSessionSessionsSessionIdPatchWithBodyWithResponse(
		t.Context(),
		uuid.MustParse(testSessionID),
		contentTypeJSON,
		strings.NewReader(`{"title":null}`),
	)
	require.NoError(t, err)
	require.NotNil(t, res.JSON200)
	assert.Nil(t, res.JSON200.Title)
	assert.Equal(t, testSessionID, res.JSON200.SessionId.String())
}

func TestSecretBindingDiscriminator(t *testing.T) {
	t.Parallel()

	var source browseruse.SecretBinding_Source

	value := testAPIKey

	inline := browseruse.InlineSecretSource{
		Type:  browseruse.InlineSecretSourceTypeInline,
		Value: &value,
	}

	require.NoError(t, source.FromInlineSecretSource(inline))

	data, err := json.Marshal(source)
	require.NoError(t, err)
	assert.JSONEq(t, `{"type":"inline","value":"EXAMPLE-DO-NOT-USE"}`,
		string(data))

	decoded, err := source.AsInlineSecretSource()
	require.NoError(t, err)
	assert.Equal(t, browseruse.InlineSecretSourceTypeInline, decoded.Type)
	require.NotNil(t, decoded.Value)
	assert.Equal(t, testAPIKey, *decoded.Value)
}

func TestCanceledRequest(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("canceled request reached the server")
	})
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	res, err := client.ListRunsRunsGetWithResponse(
		ctx, &browseruse.ListRunsRunsGetParams{},
	)
	require.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, res)
}

func TestMalformedResponse(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", contentTypeJSON)
		_, err := w.Write([]byte(`{"runs":`))
		assert.NoError(t, err)
	})

	res, err := client.ListRunsRunsGetWithResponse(
		t.Context(), &browseruse.ListRunsRunsGetParams{},
	)

	var syntaxError *json.SyntaxError

	require.ErrorAs(t, err, &syntaxError)
	assert.Nil(t, res)
}
