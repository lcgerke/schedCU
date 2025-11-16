package api

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/schedcu/v2/internal/entity"
	"github.com/schedcu/v2/internal/job"
	"github.com/schedcu/v2/internal/service"
)

// TestODSUploadHandler_Integration tests the full upload flow
func TestODSUploadHandler_Integration(t *testing.T) {
	// Setup: Create a test ODS file path
	odsFilePath := "/home/lcgerke/schedCU/cuSchedNormalized.ods"

	// Check if file exists
	_, err := os.Stat(odsFilePath)
	if err != nil {
		t.Skip("Test ODS file not found, skipping integration test")
	}

	// Create mock repository
	mockRepo := &MockScheduleVersionRepository{
		versions: make(map[string]*entity.ScheduleVersion),
	}

	// Pre-populate with a test version
	testVersionID := entity.ScheduleVersionID(uuid.New())
	testVersion := &entity.ScheduleVersion{
		ID:         testVersionID,
		HospitalID: entity.HospitalID(uuid.New()),
		Status:     entity.VersionStatusStaging,
	}
	mockRepo.versions[testVersionID.String()] = testVersion

	// Create actual version service with mock repository
	versionService := service.NewScheduleVersionService(mockRepo)

	scheduler := &job.JobScheduler{} // Mock scheduler
	services := &ServiceDeps{
		VersionService: versionService,
	}

	handlers := &Handlers{
		scheduler: scheduler,
		services:  services,
	}

	// Create test request with file upload
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// Add schedule_version_id field (use the pre-populated version ID)
	err = writer.WriteField("schedule_version_id", testVersionID.String())
	require.NoError(t, err)

	// Add file
	file, err := os.Open(odsFilePath)
	require.NoError(t, err)
	defer file.Close()

	fw, err := writer.CreateFormFile("file", "cuSchedNormalized.ods")
	require.NoError(t, err)

	_, err = io.Copy(fw, file)
	require.NoError(t, err)

	writer.Close()

	// Create HTTP request
	req := httptest.NewRequest(http.MethodPost, "/api/imports/ods/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Execute handler
	err = handlers.UploadODSFile(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify response contains expected fields
	assert.Contains(t, rec.Body.String(), "sheets_parsed")
	assert.Contains(t, rec.Body.String(), "total_assignments")
	assert.Contains(t, rec.Body.String(), "cuSchedNormalized.ods")
}

// TestODSUploadHandler_InvalidFile tests error handling
func TestODSUploadHandler_InvalidFile(t *testing.T) {
	tests := []struct {
		name           string
		filename       string
		fileContent    string
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Wrong file type",
			filename:       "schedule.xlsx",
			fileContent:    "some content",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_FILE_TYPE",
		},
		{
			name:           "Wrong file type - PDF",
			filename:       "schedule.pdf",
			fileContent:    "some content",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_FILE_TYPE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock services
			mockRepo := &MockScheduleVersionRepository{
				versions: make(map[string]*entity.ScheduleVersion),
			}
			versionService := service.NewScheduleVersionService(mockRepo)

			scheduler := &job.JobScheduler{}
			services := &ServiceDeps{
				VersionService: versionService,
			}

			handlers := &Handlers{
				scheduler: scheduler,
				services:  services,
			}

			// Create test request
			body := new(bytes.Buffer)
			writer := multipart.NewWriter(body)

			writer.WriteField("schedule_version_id", uuid.New().String())

			fw, _ := writer.CreateFormFile("file", tt.filename)
			io.WriteString(fw, tt.fileContent)
			writer.Close()

			req := httptest.NewRequest(http.MethodPost, "/api/imports/ods/upload", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			rec := httptest.NewRecorder()
			c := echo.New().NewContext(req, rec)

			err := handlers.UploadODSFile(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedCode)
		})
	}
}

// TestODSUploadHandler_MissingFile tests missing file error
func TestODSUploadHandler_MissingFile(t *testing.T) {
	mockRepo := &MockScheduleVersionRepository{
		versions: make(map[string]*entity.ScheduleVersion),
	}
	versionService := service.NewScheduleVersionService(mockRepo)

	scheduler := &job.JobScheduler{}
	services := &ServiceDeps{
		VersionService: versionService,
	}

	handlers := &Handlers{
		scheduler: scheduler,
		services:  services,
	}

	// Create request WITHOUT file
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("schedule_version_id", uuid.New().String())
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/imports/ods/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	err := handlers.UploadODSFile(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "MISSING_FILE")
}

// TestODSUploadHandler_MissingVersionID tests missing version ID error
func TestODSUploadHandler_MissingVersionID(t *testing.T) {
	// Skip if test ODS file doesn't exist
	odsFilePath := "/home/lcgerke/schedCU/cuSchedNormalized.ods"
	_, err := os.Stat(odsFilePath)
	if err != nil {
		t.Skip("Test ODS file not found, skipping test")
	}

	mockRepo := &MockScheduleVersionRepository{
		versions: make(map[string]*entity.ScheduleVersion),
	}
	versionService := service.NewScheduleVersionService(mockRepo)

	scheduler := &job.JobScheduler{}
	services := &ServiceDeps{
		VersionService: versionService,
	}

	handlers := &Handlers{
		scheduler: scheduler,
		services:  services,
	}

	// Create request without schedule_version_id, but with valid ODS file
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	file, err := os.Open(odsFilePath)
	require.NoError(t, err)
	defer file.Close()

	fw, err := writer.CreateFormFile("file", "cuSchedNormalized.ods")
	require.NoError(t, err)

	_, err = io.Copy(fw, file)
	require.NoError(t, err)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/imports/ods/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	err = handlers.UploadODSFile(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "INVALID_REQUEST")
}

// MockScheduleVersionRepository is a mock repository for testing
type MockScheduleVersionRepository struct {
	versions map[string]*entity.ScheduleVersion
}

func (m *MockScheduleVersionRepository) Create(ctx context.Context, version *entity.ScheduleVersion) error {
	m.versions[version.ID.String()] = version
	return nil
}

func (m *MockScheduleVersionRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.ScheduleVersion, error) {
	for _, v := range m.versions {
		if v.ID == entity.ScheduleVersionID(id) {
			return v, nil
		}
	}
	return nil, errors.New("not found") // Return proper error
}

func (m *MockScheduleVersionRepository) GetByHospitalAndStatus(ctx context.Context, hospitalID uuid.UUID, status entity.VersionStatus) ([]*entity.ScheduleVersion, error) {
	var result []*entity.ScheduleVersion
	for _, v := range m.versions {
		if v.HospitalID == entity.HospitalID(hospitalID) && v.Status == status {
			result = append(result, v)
		}
	}
	return result, nil
}

func (m *MockScheduleVersionRepository) GetActiveVersion(ctx context.Context, hospitalID uuid.UUID, date time.Time) (*entity.ScheduleVersion, error) {
	return nil, errors.New("not found")
}

func (m *MockScheduleVersionRepository) Update(ctx context.Context, version *entity.ScheduleVersion) error {
	m.versions[version.ID.String()] = version
	return nil
}

func (m *MockScheduleVersionRepository) Delete(ctx context.Context, id uuid.UUID, deleterID uuid.UUID) error {
	return nil
}

func (m *MockScheduleVersionRepository) ListByHospital(ctx context.Context, hospitalID uuid.UUID) ([]*entity.ScheduleVersion, error) {
	var result []*entity.ScheduleVersion
	for _, v := range m.versions {
		if v.HospitalID == entity.HospitalID(hospitalID) {
			result = append(result, v)
		}
	}
	return result, nil
}

func (m *MockScheduleVersionRepository) Count(ctx context.Context) (int64, error) {
	return int64(len(m.versions)), nil
}

