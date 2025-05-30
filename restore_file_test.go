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

func TestRestore_File_List(t *testing.T) {
	testService := goslide.NewService("fakeToken",
		goslide.WithCustomRoundtripper(
			roundtripper.NetworkQueue(
				t,
				[]roundtripper.TestRoundTripFunc{
					roundtripper.ServeAndValidate(
						t,
						&roundtripper.TestResponseFile{
							StatusCode: http.StatusOK,
							FilePath:   "testdata/responses/restore_file/list_page1_200.json",
						},
						roundtripper.ExpectedTestRequest{
							Method: http.MethodGet,
							Path:   "/v1/restore/file",
							Query:  url.Values{},
						},
					),
					roundtripper.ServeAndValidate(
						t,
						&roundtripper.TestResponseFile{
							StatusCode: http.StatusOK,
							FilePath:   "testdata/responses/restore_file/list_page2_200.json",
						},
						roundtripper.ExpectedTestRequest{
							Method: http.MethodGet,
							Path:   "/v1/restore/file",
							Query: url.Values{
								"offset": []string{"1"},
							},
						},
					),
				},
			),
		),
	)

	actual := []goslide.FileRestore{}

	ctx := context.Background()
	if err := testService.FileRestores().List(ctx,
		func(response goslide.ListResponse[goslide.FileRestore]) error {
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

func TestRestore_File_Get(t *testing.T) {
	fileRestoreID := "fr_0123456789ab"

	testService := goslide.NewService("fakeToken",
		goslide.WithCustomRoundtripper(
			roundtripper.NetworkQueue(
				t,
				[]roundtripper.TestRoundTripFunc{
					roundtripper.ServeAndValidate(
						t,
						&roundtripper.TestResponseFile{
							StatusCode: http.StatusOK,
							FilePath:   "testdata/responses/restore_file/get_200.json",
						},
						roundtripper.ExpectedTestRequest{
							Method: http.MethodGet,
							Path:   "/v1/restore/file/" + fileRestoreID,
							Query:  url.Values{},
						},
					),
				},
			),
		),
	)

	ctx := context.Background()
	actual, err := testService.FileRestores().Get(ctx, fileRestoreID)
	if err != nil {
		t.Fatal(err)
	}

	expected := goslide.FileRestore{
		AgentID:       "a_0123456789ab",
		CreatedAt:     generateRFC3389FromString(t, "2024-08-23T01:25:08Z"),
		DeviceID:      "d_0123456789ab",
		ExpiresAt:     generateRFC3389FromString(t, "2024-08-23T01:25:08Z"),
		FileRestoreID: fileRestoreID,
		SnapshotID:    "s_0123456789ab",
	}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Fatalf("%s Returned struct mismatch (-want +got):\n%s", t.Name(), diff)
	}
}

func TestRestore_File_Delete(t *testing.T) {
	fileRestoreID := "fr_0123456789ab"

	testService := goslide.NewService("fakeToken",
		goslide.WithCustomRoundtripper(
			roundtripper.NetworkQueue(
				t,
				[]roundtripper.TestRoundTripFunc{
					roundtripper.ServeAndValidate(
						t,
						&roundtripper.TestResponseNoContent{
							StatusCode: http.StatusNoContent,
						},
						roundtripper.ExpectedTestRequest{
							Method: http.MethodDelete,
							Path:   "/v1/restore/file/" + fileRestoreID,
							Query:  url.Values{},
						},
					),
				},
			),
		),
	)

	ctx := context.Background()
	if err := testService.FileRestores().Delete(ctx, fileRestoreID); err != nil {
		t.Fatal(err)
	}
}

func TestRestore_File_Create(t *testing.T) {
	testService := goslide.NewService("fakeToken",
		goslide.WithCustomRoundtripper(
			roundtripper.NetworkQueue(
				t,
				[]roundtripper.TestRoundTripFunc{
					roundtripper.ServeAndValidate(
						t,
						&roundtripper.TestResponseFile{
							StatusCode: http.StatusCreated,
							FilePath:   "testdata/responses/restore_file/create_201.json",
						},
						roundtripper.ExpectedTestRequest{
							Method: http.MethodPost,
							Path:   "/v1/restore/file",
							Query:  url.Values{},
							Validator: func(r *http.Request) error {
								expectedBody, err := os.ReadFile("testdata/requests/restore_file/create_201.json")
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
	actual, err := testService.FileRestores().Create(ctx, goslide.FileRestorePayload{
		DeviceID:   "d_0123456789ab",
		SnapshotID: "s_0123456789ab",
	})
	if err != nil {
		t.Fatal(err)
	}

	expected := goslide.FileRestore{
		AgentID:       "a_0123456789ab",
		CreatedAt:     generateRFC3389FromString(t, "2024-08-23T01:25:08Z"),
		DeviceID:      "d_0123456789ab",
		ExpiresAt:     generateRFC3389FromString(t, "2024-08-23T01:25:08Z"),
		FileRestoreID: "fr_0123456789ab",
		SnapshotID:    "s_0123456789ab",
	}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Fatalf("%s Returned struct mismatch (-want +got):\n%s", t.Name(), diff)
	}
}

func TestRestore_File_Browse(t *testing.T) {
	fileRestoreID := "fr_0123456789ab"
	testService := goslide.NewService("fakeToken",
		goslide.WithCustomRoundtripper(
			roundtripper.NetworkQueue(
				t,
				[]roundtripper.TestRoundTripFunc{
					roundtripper.ServeAndValidate(
						t,
						&roundtripper.TestResponseFile{
							StatusCode: http.StatusOK,
							FilePath:   "testdata/responses/restore_file/browse_page1_200.json",
						},
						roundtripper.ExpectedTestRequest{
							Method: http.MethodGet,
							Path:   "/v1/restore/file/" + fileRestoreID + "/browse",
							Query:  url.Values{},
						},
					),
					roundtripper.ServeAndValidate(
						t,
						&roundtripper.TestResponseFile{
							StatusCode: http.StatusOK,
							FilePath:   "testdata/responses/restore_file/browse_page2_200.json",
						},
						roundtripper.ExpectedTestRequest{
							Method: http.MethodGet,
							Path:   "/v1/restore/file/" + fileRestoreID + "/browse",
							Query: url.Values{
								"offset": []string{"1"},
							},
						},
					),
				},
			),
		),
	)

	actual := []goslide.FileRestoreData{}

	ctx := context.Background()
	if err := testService.FileRestores().Browse(
		ctx,
		fileRestoreID,
		func(response goslide.ListResponse[goslide.FileRestoreData]) error {
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
