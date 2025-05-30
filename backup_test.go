package goslide_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/equalsgibson/goslide"
	"github.com/equalsgibson/goslide/internal/roundtripper"
	"github.com/google/go-cmp/cmp"
)

func TestBackup_List(t *testing.T) {
	testService := goslide.NewService("fakeToken",
		goslide.WithCustomRoundtripper(
			roundtripper.NetworkQueue(
				t,
				[]roundtripper.TestRoundTripFunc{
					roundtripper.ServeAndValidate(
						t,
						&roundtripper.TestResponseFile{
							StatusCode: http.StatusOK,
							FilePath:   "testdata/responses/backup/list_page1_200.json",
						},
						roundtripper.ExpectedTestRequest{
							Method: http.MethodGet,
							Path:   "/v1/backup",
							Query:  url.Values{},
						},
					),
					roundtripper.ServeAndValidate(
						t,
						&roundtripper.TestResponseFile{
							StatusCode: http.StatusOK,
							FilePath:   "testdata/responses/backup/list_page2_200.json",
						},
						roundtripper.ExpectedTestRequest{
							Method: http.MethodGet,
							Path:   "/v1/backup",
							Query: url.Values{
								"offset": []string{"1"},
							},
						},
					),
				},
			),
		),
	)

	actual := []goslide.Backup{}

	ctx := context.Background()
	if err := testService.Backups().List(ctx,
		func(response goslide.ListResponse[goslide.Backup]) error {
			actual = append(actual, response.Data...)

			return nil
		},
	); err != nil {
		t.Fatal(err)
	}

	if len(actual) != 2 {
		t.Fatal(actual)
	}
}

func TestBackup_StartBackup(t *testing.T) {
	testService := goslide.NewService("fakeToken",
		goslide.WithCustomRoundtripper(
			roundtripper.NetworkQueue(
				t,
				[]roundtripper.TestRoundTripFunc{
					roundtripper.ServeAndValidate(
						t,
						&roundtripper.TestResponseNoContent{
							StatusCode: http.StatusAccepted,
						},
						roundtripper.ExpectedTestRequest{
							Method: http.MethodPost,
							Path:   "/v1/backup",
							Query:  url.Values{},
							Validator: func(r *http.Request) error {
								expectedBody, err := os.ReadFile("testdata/requests/backup/start_backup_202.json")
								if err != nil {
									return fmt.Errorf("error during test setup - could not read file: %w", err)
								}

								actualBody, err := io.ReadAll(r.Body)
								if err != nil {
									return fmt.Errorf("error during test setup - could not read request body: %w", err)
								}
								r.Body = io.NopCloser(bytes.NewBuffer(actualBody))

								var actualBodyFormatted bytes.Buffer
								if err := json.Indent(&actualBodyFormatted, actualBody, "", "    "); err != nil {
									return fmt.Errorf("error during test setup - could not format request body: %w", err)
								}

								if diff := cmp.Diff(string(expectedBody), actualBodyFormatted.String()); diff != "" {
									t.Fatalf("%s Expected Request Body mismatch (-want +got):\n%s", t.Name(), diff)
								}

								return nil
							},
						},
					),
				},
			),
		),
	)

	ctx := context.Background()

	if err := testService.Backups().StartBackup(ctx, "a_0123456789ab"); err != nil {
		t.Fatal(err)
	}
}

func TestBackup_Get(t *testing.T) {
	agentID := "al_0123456789ab"
	testService := goslide.NewService("fakeToken",
		goslide.WithCustomRoundtripper(
			roundtripper.NetworkQueue(
				t,
				[]roundtripper.TestRoundTripFunc{
					roundtripper.ServeAndValidate(
						t,
						&roundtripper.TestResponseFile{
							StatusCode: http.StatusOK,
							FilePath:   "testdata/responses/backup/get_200.json",
						},
						roundtripper.ExpectedTestRequest{
							Method: http.MethodGet,
							Path:   "/v1/backup/" + agentID,
							Query:  url.Values{},
						},
					),
				},
			),
		),
	)

	ctx := context.Background()
	actual, err := testService.Backups().Get(ctx, agentID)
	if err != nil {
		t.Fatal(err)
	}

	expected := goslide.Backup{
		AgentID:      "a_0123456789ab",
		BackupID:     "b_0123456789ab",
		EndedAt:      generateRFC3389FromString(t, "2024-08-23T01:40:08Z"),
		ErrorCode:    1,
		ErrorMessage: "string",
		SnapshotID:   "s_0123456789ab",
		StartedAt:    generateRFC3389FromString(t, "2024-08-23T01:25:08Z"),
		Status:       goslide.BackupStatus_SUCCEEDED,
	}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Fatalf("%s Returned struct mismatch (-want +got):\n%s", t.Name(), diff)
	}
}
