package output

import (
	"strings"
	"testing"
	"time"

	"synkronus/internal/domain"
	"synkronus/internal/domain/storage"
)

func TestBucketListView_RenderTable(t *testing.T) {
	buckets := BucketListView{
		{
			Name:         "my-bucket",
			Provider:     domain.GCP,
			Location:     "US-CENTRAL1",
			StorageClass: "STANDARD",
			UsageBytes:   1048576, // 1 MB
			CreatedAt:    time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	result := buckets.RenderTable()

	expectedSubstrings := []string{
		"BUCKET NAME", "PROVIDER", "LOCATION", "USAGE", "STORAGE CLASS", "CREATED",
		"my-bucket", "GCP", "US-CENTRAL1", "1.0 MB", "STANDARD", "2025-01-15",
	}

	for _, s := range expectedSubstrings {
		if !strings.Contains(result, s) {
			t.Errorf("expected output to contain %q, got:\n%s", s, result)
		}
	}
}

func TestBucketListView_Empty(t *testing.T) {
	buckets := BucketListView{}

	result := buckets.RenderTable()

	headers := []string{"BUCKET NAME", "PROVIDER", "LOCATION", "USAGE", "STORAGE CLASS", "CREATED"}
	for _, h := range headers {
		if !strings.Contains(result, h) {
			t.Errorf("empty list should still contain header %q, got:\n%s", h, result)
		}
	}
}

func TestBucketDetailView_RenderTable(t *testing.T) {
	bucket := storage.Bucket{
		Name:         "detail-bucket",
		Provider:     domain.GCP,
		Location:     "US-EAST1",
		LocationType: "region",
		StorageClass: "NEARLINE",
		UsageBytes:   5368709120, // 5 GB
		CreatedAt:    time.Date(2025, 3, 1, 10, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2025, 3, 15, 12, 0, 0, 0, time.UTC),
		Autoclass:    &storage.Autoclass{Enabled: true},
		Versioning:   &storage.Versioning{Enabled: true},
		Encryption:   &storage.Encryption{KmsKeyName: "projects/my-proj/locations/us/keyRings/kr/cryptoKeys/key1"},
		Labels:       map[string]string{"env": "prod", "team": "data"},
		IAMPolicy: &storage.IAMPolicy{
			Bindings: []storage.IAMBinding{
				{Role: "roles/storage.admin", Principals: []string{"user:admin@example.com"}},
			},
		},
		UniformBucketLevelAccess: &storage.UniformBucketLevelAccess{Enabled: true},
		PublicAccessPrevention:   "enforced",
		SoftDeletePolicy:         &storage.SoftDeletePolicy{RetentionDuration: 7 * 24 * time.Hour},
		RetentionPolicy:          &storage.RetentionPolicy{RetentionPeriod: 30 * 24 * time.Hour, IsLocked: true},
		LifecycleRules: []storage.LifecycleRule{
			{Action: "Delete", Condition: storage.LifecycleCondition{Age: 365}},
		},
	}

	view := BucketDetailView{bucket}
	result := view.RenderTable()

	expectedSections := []string{
		"Bucket: detail-bucket",
		"-- Overview --",
		"-- Access Control & Logging --",
		"-- Data Protection --",
		"-- Lifecycle Rules --",
		"-- Labels --",
	}

	for _, section := range expectedSections {
		if !strings.Contains(result, section) {
			t.Errorf("expected output to contain section %q, got:\n%s", section, result)
		}
	}

	// Verify specific field values
	expectedValues := []string{
		"GCP", "US-EAST1", "region", "NEARLINE", "Enabled", // Autoclass
		"projects/my-proj/locations/us/keyRings/kr/cryptoKeys/key1", // CMEK
		"env", "prod", "team", "data", // Labels
		"roles/storage.admin", "user:admin@example.com", // IAM
		"Locked", // Retention
		"Delete", "Age > 365 days", // Lifecycle
	}

	for _, v := range expectedValues {
		if !strings.Contains(result, v) {
			t.Errorf("expected output to contain %q, got:\n%s", v, result)
		}
	}
}

func TestBucketDetailView_NilOptionalFields(t *testing.T) {
	bucket := storage.Bucket{
		Name:         "minimal-bucket",
		Provider:     domain.GCP,
		Location:     "US",
		StorageClass: "STANDARD",
		UsageBytes:   0,
		CreatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	view := BucketDetailView{bucket}

	// Should not panic with nil optional fields
	result := view.RenderTable()

	if !strings.Contains(result, "minimal-bucket") {
		t.Errorf("expected bucket name in output, got:\n%s", result)
	}

	// Lifecycle and Labels sections should be absent when empty
	if strings.Contains(result, "-- Lifecycle Rules --") {
		t.Error("lifecycle section should not appear when no rules exist")
	}
	if strings.Contains(result, "-- Labels --") {
		t.Error("labels section should not appear when no labels exist")
	}
}

func TestObjectListView_RenderTable(t *testing.T) {
	objectList := storage.ObjectList{
		BucketName: "my-bucket",
		Prefix:     "data/",
		Objects: []storage.Object{
			{
				Key:          "data/file1.csv",
				Size:         2048,
				StorageClass: "STANDARD",
				LastModified: time.Date(2025, 2, 10, 8, 30, 0, 0, time.UTC),
			},
		},
		CommonPrefixes: []string{"data/subdir/"},
	}

	view := ObjectListView{objectList}
	result := view.RenderTable()

	expectedSubstrings := []string{
		"my-bucket",
		"data/",
		"KEY", "SIZE", "STORAGE CLASS", "LAST MODIFIED",
		"data/subdir/", "(DIR)",
		"data/file1.csv", "2.0 KB", "STANDARD",
	}

	for _, s := range expectedSubstrings {
		if !strings.Contains(result, s) {
			t.Errorf("expected output to contain %q, got:\n%s", s, result)
		}
	}
}

func TestObjectListView_Empty(t *testing.T) {
	objectList := storage.ObjectList{
		BucketName: "empty-bucket",
	}

	view := ObjectListView{objectList}
	result := view.RenderTable()

	if !strings.Contains(result, "No objects or directories found") {
		t.Errorf("expected empty message, got:\n%s", result)
	}
}

func TestObjectDetailView_RenderTable(t *testing.T) {
	object := storage.Object{
		Key:          "data/report.pdf",
		Bucket:       "my-bucket",
		Provider:     domain.GCP,
		Size:         1048576,
		StorageClass: "STANDARD",
		LastModified: time.Date(2025, 3, 1, 12, 0, 0, 0, time.UTC),
		CreatedAt:    time.Date(2025, 3, 1, 12, 0, 0, 0, time.UTC),
		ETag:         "abc123",
		ContentType:  "application/pdf",
		Encryption:   &storage.Encryption{KmsKeyName: "projects/p/locations/l/keyRings/kr/cryptoKeys/k", Algorithm: "AES256"},
		CRC32C:       "AABBCC==",
		Generation:   12345,
		Metageneration: 2,
		Metadata:     map[string]string{"author": "test-user"},
	}

	view := ObjectDetailView{object}
	result := view.RenderTable()

	expectedSections := []string{
		"Object: data/report.pdf",
		"-- Overview --",
		"-- HTTP Headers --",
		"-- User-Defined Metadata --",
	}

	for _, section := range expectedSections {
		if !strings.Contains(result, section) {
			t.Errorf("expected output to contain section %q, got:\n%s", section, result)
		}
	}

	expectedValues := []string{
		"application/pdf",
		"AES256",
		"projects/p/locations/l/keyRings/kr/cryptoKeys/k",
		"CRC32C",
		"AABBCC==",
		"Generation",
		"12345",
		"author", "test-user",
	}

	for _, v := range expectedValues {
		if !strings.Contains(result, v) {
			t.Errorf("expected output to contain %q, got:\n%s", v, result)
		}
	}
}

func TestBucketDetailView_UBLADisabled_ShowsACLs(t *testing.T) {
	bucket := storage.Bucket{
		Name:                     "acl-bucket",
		Provider:                 domain.GCP,
		Location:                 "US",
		StorageClass:             "STANDARD",
		CreatedAt:                time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:                time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UniformBucketLevelAccess: &storage.UniformBucketLevelAccess{Enabled: false},
		PublicAccessPrevention:   "inherited",
		ACLs: []storage.ACLRule{
			{Entity: "user:admin@example.com", Role: "OWNER"},
			{Entity: "allUsers", Role: "READER"},
		},
	}
	view := BucketDetailView{bucket}
	result := view.RenderTable()

	if !strings.Contains(result, "Fine-grained") {
		t.Errorf("expected UBLA disabled to show fine-grained text, got:\n%s", result)
	}
	if !strings.Contains(result, "user:admin@example.com") {
		t.Errorf("expected ACL entity in output, got:\n%s", result)
	}
	if !strings.Contains(result, "allUsers") {
		t.Errorf("expected allUsers ACL in output, got:\n%s", result)
	}
}

func TestBucketDetailView_NilIAMPolicy(t *testing.T) {
	bucket := storage.Bucket{
		Name:         "no-iam-bucket",
		Provider:     domain.GCP,
		Location:     "US",
		StorageClass: "STANDARD",
		CreatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		IAMPolicy:    nil,
	}
	view := BucketDetailView{bucket}
	result := view.RenderTable()

	if !strings.Contains(result, "Could not retrieve IAM policy") {
		t.Errorf("expected 'Could not retrieve' message for nil IAM policy, got:\n%s", result)
	}
}

func TestBucketDetailView_LifecycleMultipleConditions(t *testing.T) {
	bucket := storage.Bucket{
		Name:         "lifecycle-bucket",
		Provider:     domain.GCP,
		Location:     "US",
		StorageClass: "STANDARD",
		CreatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		LifecycleRules: []storage.LifecycleRule{
			{
				Action: "Delete",
				Condition: storage.LifecycleCondition{
					Age:                 90,
					NumNewerVersions:    3,
					MatchesStorageClass: []string{"NEARLINE", "COLDLINE"},
					CreatedBefore:       time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
				},
			},
		},
	}
	view := BucketDetailView{bucket}
	result := view.RenderTable()

	expectedConditions := []string{
		"Age > 90 days",
		"NumNewerVersions = 3",
		"StorageClass IN (NEARLINE, COLDLINE)",
		"CreatedBefore = 2024-06-01",
	}
	for _, cond := range expectedConditions {
		if !strings.Contains(result, cond) {
			t.Errorf("expected lifecycle condition %q in output, got:\n%s", cond, result)
		}
	}
}

func TestObjectDetailView_AllHTTPHeaders(t *testing.T) {
	object := storage.Object{
		Key:                "headers.txt",
		Bucket:             "my-bucket",
		Provider:           domain.GCP,
		CreatedAt:          time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		ContentType:        "text/plain",
		ContentEncoding:    "gzip",
		ContentLanguage:    "en-US",
		CacheControl:       "max-age=3600",
		ContentDisposition: "attachment; filename=headers.txt",
	}
	view := ObjectDetailView{object}
	result := view.RenderTable()

	expectedHeaders := []string{
		"Content-Type", "text/plain",
		"Content-Encoding", "gzip",
		"Content-Language", "en-US",
		"Cache-Control", "max-age=3600",
		"Content-Disposition", "attachment",
	}
	for _, h := range expectedHeaders {
		if !strings.Contains(result, h) {
			t.Errorf("expected HTTP header %q in output, got:\n%s", h, result)
		}
	}
}

func TestObjectDetailView_ZeroCreatedAt(t *testing.T) {
	object := storage.Object{
		Key:      "no-time.txt",
		Bucket:   "my-bucket",
		Provider: domain.GCP,
	}
	view := ObjectDetailView{object}
	result := view.RenderTable()

	if !strings.Contains(result, "N/A") {
		t.Errorf("expected 'N/A' for zero CreatedAt, got:\n%s", result)
	}
}

func TestBucketListView_ZeroCreatedAt(t *testing.T) {
	buckets := BucketListView{
		{
			Name:         "aws-bucket",
			Provider:     domain.AWS,
			Location:     "us-east-1",
			StorageClass: "STANDARD",
			UsageBytes:   -1,
			// CreatedAt is zero value
		},
	}
	result := buckets.RenderTable()
	if strings.Contains(result, "0001") {
		t.Errorf("expected zero time to render as N/A, got:\n%s", result)
	}
	if !strings.Contains(result, "N/A") {
		t.Errorf("expected 'N/A' for zero CreatedAt, got:\n%s", result)
	}
}

func TestBucketDetailView_ZeroTimestamps(t *testing.T) {
	bucket := storage.Bucket{
		Name:         "aws-bucket",
		Provider:     domain.AWS,
		Location:     "us-east-1",
		StorageClass: "STANDARD",
		UsageBytes:   -1,
		// Both CreatedAt and UpdatedAt are zero
	}
	view := BucketDetailView{bucket}
	result := view.RenderTable()

	// Should contain N/A for both timestamps, not "0001-01-01"
	if strings.Contains(result, "0001") {
		t.Errorf("expected zero times to render as N/A, got:\n%s", result)
	}
}

func TestBucketDetailView_AWSPolicyStatements(t *testing.T) {
	bucket := storage.Bucket{
		Name:         "policy-bucket",
		Provider:     domain.AWS,
		Location:     "us-east-1",
		StorageClass: "STANDARD",
		IAMPolicy: &storage.IAMPolicy{
			Statements: []storage.PolicyStatement{
				{
					Effect:     "Allow",
					Principals: []string{"*"},
					Actions:    []string{"s3:GetObject"},
					Resources:  []string{"arn:aws:s3:::policy-bucket/*"},
				},
			},
		},
	}
	view := BucketDetailView{bucket}
	result := view.RenderTable()

	expectedValues := []string{
		"Bucket Policy Statements",
		"Allow",
		"s3:GetObject",
		"arn:aws:s3:::policy-bucket/*",
	}
	for _, v := range expectedValues {
		if !strings.Contains(result, v) {
			t.Errorf("expected output to contain %q, got:\n%s", v, result)
		}
	}
}

func TestBucketDetailView_AWSPolicyStatementsWithConditions(t *testing.T) {
	bucket := storage.Bucket{
		Name:         "cond-bucket",
		Provider:     domain.AWS,
		Location:     "us-east-1",
		StorageClass: "STANDARD",
		IAMPolicy: &storage.IAMPolicy{
			Statements: []storage.PolicyStatement{
				{
					Effect:     "Allow",
					Principals: []string{"*"},
					Actions:    []string{"s3:GetObject"},
					Resources:  []string{"arn:aws:s3:::cond-bucket/*"},
					Conditions: map[string]map[string][]string{
						"StringLike": {"s3:prefix": {"home/", "home/*"}},
					},
				},
			},
		},
	}
	view := BucketDetailView{bucket}
	result := view.RenderTable()

	if !strings.Contains(result, "1 condition(s) present") {
		t.Errorf("expected condition count note, got:\n%s", result)
	}
}

func TestBucketDetailView_NilUBLA(t *testing.T) {
	bucket := storage.Bucket{
		Name:                     "aws-bucket",
		Provider:                 domain.AWS,
		Location:                 "us-east-1",
		StorageClass:             "STANDARD",
		UniformBucketLevelAccess: nil,
	}
	view := BucketDetailView{bucket}
	result := view.RenderTable()

	// Should not contain UBLA-related text
	if strings.Contains(result, "Uniform") {
		t.Errorf("expected no UBLA text for nil, got:\n%s", result)
	}
	// Should not crash — this is the main assertion
}

func TestObjectDetailView_AWSVersionID(t *testing.T) {
	object := storage.Object{
		Key:       "versioned.txt",
		Bucket:    "my-bucket",
		Provider:  domain.AWS,
		VersionID: "abc123-version-id",
	}
	view := ObjectDetailView{object}
	result := view.RenderTable()

	if !strings.Contains(result, "Version ID") {
		t.Errorf("expected 'Version ID' row, got:\n%s", result)
	}
	if !strings.Contains(result, "abc123-version-id") {
		t.Errorf("expected version ID value, got:\n%s", result)
	}
}

func TestBucketDetailView_LifecycleWithPrefix(t *testing.T) {
	bucket := storage.Bucket{
		Name:         "prefix-bucket",
		Provider:     domain.AWS,
		Location:     "us-east-1",
		StorageClass: "STANDARD",
		LifecycleRules: []storage.LifecycleRule{
			{
				Action: "Delete",
				Condition: storage.LifecycleCondition{
					Age:    90,
					Prefix: "logs/",
				},
			},
		},
	}
	view := BucketDetailView{bucket}
	result := view.RenderTable()

	if !strings.Contains(result, "Prefix = logs/") {
		t.Errorf("expected 'Prefix = logs/' condition, got:\n%s", result)
	}
}

func TestBucketDetailView_IAMPolicyWithConditions(t *testing.T) {
	bucket := storage.Bucket{
		Name:         "cond-bucket",
		Provider:     domain.GCP,
		Location:     "US",
		StorageClass: "STANDARD",
		CreatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		IAMPolicy: &storage.IAMPolicy{
			Bindings: []storage.IAMBinding{
				{
					Role:       "roles/storage.objectViewer",
					Principals: []string{"user:a@example.com", "user:b@example.com"},
					Condition: &storage.IAMCondition{
						Title:      "Business hours only",
						Expression: "request.time.getHours('America/Chicago') >= 9",
					},
				},
			},
		},
	}
	view := BucketDetailView{bucket}
	result := view.RenderTable()

	if !strings.Contains(result, "Conditions:") {
		t.Errorf("expected 'Conditions:' header, got:\n%s", result)
	}
	if !strings.Contains(result, "Business hours only") {
		t.Errorf("expected condition title in output, got:\n%s", result)
	}
	if !strings.Contains(result, "request.time.getHours") {
		t.Errorf("expected condition expression in output, got:\n%s", result)
	}
	if !strings.Contains(result, "user:a@example.com") || !strings.Contains(result, "user:b@example.com") {
		t.Errorf("expected both principals in output, got:\n%s", result)
	}
}

func TestBucketDetailView_IAMConditionWithDescription(t *testing.T) {
	bucket := storage.Bucket{
		Name:         "desc-cond-bucket",
		Provider:     domain.GCP,
		Location:     "US",
		StorageClass: "STANDARD",
		CreatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		IAMPolicy: &storage.IAMPolicy{
			Bindings: []storage.IAMBinding{
				{
					Role:       "roles/storage.admin",
					Principals: []string{"user:admin@example.com"},
					Condition: &storage.IAMCondition{
						Title:       "Temporary access",
						Description: "Granted for incident response until 2026-05-01",
						Expression:  "request.time < timestamp('2026-05-01T00:00:00Z')",
					},
				},
			},
		},
	}
	view := BucketDetailView{bucket}
	result := view.RenderTable()

	if !strings.Contains(result, "Temporary access") {
		t.Errorf("expected condition title, got:\n%s", result)
	}
	if !strings.Contains(result, "Granted for incident response") {
		t.Errorf("expected condition description, got:\n%s", result)
	}
	if !strings.Contains(result, "request.time < timestamp") {
		t.Errorf("expected condition expression, got:\n%s", result)
	}
}

func TestBucketDetailView_IAMMixedConditions(t *testing.T) {
	bucket := storage.Bucket{
		Name:         "mixed-bucket",
		Provider:     domain.GCP,
		Location:     "US",
		StorageClass: "STANDARD",
		CreatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		IAMPolicy: &storage.IAMPolicy{
			Bindings: []storage.IAMBinding{
				{
					Role:       "roles/storage.admin",
					Principals: []string{"user:admin@example.com"},
				},
				{
					Role:       "roles/storage.objectViewer",
					Principals: []string{"user:viewer@example.com"},
					Condition: &storage.IAMCondition{
						Title:      "IP restricted",
						Expression: "request.auth.claims.source_ip == '10.0.0.1'",
					},
				},
			},
		},
	}
	view := BucketDetailView{bucket}
	result := view.RenderTable()

	if !strings.Contains(result, "roles/storage.admin") {
		t.Errorf("expected unconditional role, got:\n%s", result)
	}
	if !strings.Contains(result, "roles/storage.objectViewer") {
		t.Errorf("expected conditional role, got:\n%s", result)
	}
	if !strings.Contains(result, "IP restricted") {
		t.Errorf("expected condition title for objectViewer, got:\n%s", result)
	}
	if strings.Contains(result, "roles/storage.admin -") {
		t.Errorf("unconditional binding should not have condition annotation, got:\n%s", result)
	}
}

func TestBucketDetailView_IAMDuplicateRoleConditions(t *testing.T) {
	bucket := storage.Bucket{
		Name:         "dup-role-bucket",
		Provider:     domain.GCP,
		Location:     "US",
		StorageClass: "STANDARD",
		CreatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		IAMPolicy: &storage.IAMPolicy{
			Bindings: []storage.IAMBinding{
				{
					Role:       "roles/storage.objectViewer",
					Principals: []string{"user:a@example.com"},
					Condition: &storage.IAMCondition{
						Title:      "Business hours",
						Expression: "request.time.getHours('America/Chicago') >= 9",
					},
				},
				{
					Role:       "roles/storage.objectViewer",
					Principals: []string{"user:b@example.com"},
					Condition: &storage.IAMCondition{
						Title:      "Weekend only",
						Expression: "request.time.getDayOfWeek() >= 6",
					},
				},
			},
		},
	}
	view := BucketDetailView{bucket}
	result := view.RenderTable()

	// Both condition annotations should include the first principal for disambiguation
	if !strings.Contains(result, "user:a@example.com) - Business hours") {
		t.Errorf("expected first principal in condition annotation, got:\n%s", result)
	}
	if !strings.Contains(result, "user:b@example.com) - Weekend only") {
		t.Errorf("expected second principal in condition annotation, got:\n%s", result)
	}
}
