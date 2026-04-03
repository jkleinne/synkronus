package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func TestMapTags_Empty(t *testing.T) {
	result := mapTags(nil)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestMapTags_Multiple(t *testing.T) {
	tags := []types.Tag{
		{Key: strPtr("env"), Value: strPtr("prod")},
		{Key: strPtr("team"), Value: strPtr("data")},
	}
	result := mapTags(tags)
	if len(result) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(result))
	}
	if result["env"] != "prod" {
		t.Errorf("expected env=prod, got %s", result["env"])
	}
	if result["team"] != "data" {
		t.Errorf("expected team=data, got %s", result["team"])
	}
}

func TestMapACLGrants_CanonicalUser(t *testing.T) {
	owner := &types.Owner{ID: strPtr("owner-id"), DisplayName: strPtr("owner-display")}
	grants := []types.Grant{
		{
			Grantee:    &types.Grantee{Type: types.TypeCanonicalUser, DisplayName: strPtr("admin"), ID: strPtr("admin-id")},
			Permission: types.PermissionFullControl,
		},
	}
	result := mapACLGrants(owner, grants)
	if len(result) != 1 {
		t.Fatalf("expected 1 ACL rule, got %d", len(result))
	}
	if result[0].Entity != "admin" {
		t.Errorf("expected entity 'admin', got %q", result[0].Entity)
	}
	if result[0].Role != string(types.PermissionFullControl) {
		t.Errorf("expected role %q, got %q", types.PermissionFullControl, result[0].Role)
	}
}

func TestMapACLGrants_GroupURI(t *testing.T) {
	grants := []types.Grant{
		{
			Grantee:    &types.Grantee{Type: types.TypeGroup, URI: strPtr("http://acs.amazonaws.com/groups/global/AllUsers")},
			Permission: types.PermissionRead,
		},
	}
	result := mapACLGrants(nil, grants)
	if len(result) != 1 {
		t.Fatalf("expected 1 ACL rule, got %d", len(result))
	}
	if result[0].Entity != "http://acs.amazonaws.com/groups/global/AllUsers" {
		t.Errorf("unexpected entity: %q", result[0].Entity)
	}
}

func TestMapACLGrants_Empty(t *testing.T) {
	result := mapACLGrants(nil, nil)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestMapLifecycleRules_Expiration(t *testing.T) {
	days := int32(90)
	rules := []types.LifecycleRule{
		{
			Status:     types.ExpirationStatusEnabled,
			Expiration: &types.LifecycleExpiration{Days: &days},
			Filter:     &types.LifecycleRuleFilter{Prefix: strPtr("logs/")},
		},
	}
	result := mapLifecycleRules(rules)
	if len(result) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(result))
	}
	if result[0].Action != "Delete" {
		t.Errorf("expected action 'Delete', got %q", result[0].Action)
	}
	if result[0].Condition.Age != 90 {
		t.Errorf("expected age 90, got %d", result[0].Condition.Age)
	}
	if result[0].Condition.Prefix != "logs/" {
		t.Errorf("expected prefix 'logs/', got %q", result[0].Condition.Prefix)
	}
}

func TestMapLifecycleRules_Transition(t *testing.T) {
	days := int32(30)
	rules := []types.LifecycleRule{
		{
			Status: types.ExpirationStatusEnabled,
			Transitions: []types.Transition{
				{Days: &days, StorageClass: types.TransitionStorageClassGlacier},
			},
		},
	}
	result := mapLifecycleRules(rules)
	if len(result) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(result))
	}
	if result[0].Action != "Transition to GLACIER" {
		t.Errorf("expected action 'Transition to GLACIER', got %q", result[0].Action)
	}
	if result[0].Condition.Age != 30 {
		t.Errorf("expected age 30, got %d", result[0].Condition.Age)
	}
}

func TestMapLifecycleRules_DisabledSkipped(t *testing.T) {
	days := int32(90)
	rules := []types.LifecycleRule{
		{
			Status:     types.ExpirationStatusDisabled,
			Expiration: &types.LifecycleExpiration{Days: &days},
		},
	}
	result := mapLifecycleRules(rules)
	if len(result) != 0 {
		t.Errorf("expected 0 rules for disabled, got %d", len(result))
	}
}

func TestMapLifecycleRules_Empty(t *testing.T) {
	result := mapLifecycleRules(nil)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestMapEncryption_SSES3(t *testing.T) {
	rules := []types.ServerSideEncryptionRule{
		{ApplyServerSideEncryptionByDefault: &types.ServerSideEncryptionByDefault{SSEAlgorithm: types.ServerSideEncryptionAes256}},
	}
	result := mapEncryption(rules)
	if result == nil {
		t.Fatal("expected non-nil encryption")
	}
	if result.Algorithm != string(types.ServerSideEncryptionAes256) {
		t.Errorf("expected AES256, got %q", result.Algorithm)
	}
	if result.KmsKeyName != "" {
		t.Errorf("expected empty KMS key, got %q", result.KmsKeyName)
	}
}

func TestMapEncryption_SSEKMS(t *testing.T) {
	rules := []types.ServerSideEncryptionRule{
		{ApplyServerSideEncryptionByDefault: &types.ServerSideEncryptionByDefault{
			SSEAlgorithm:   types.ServerSideEncryptionAwsKms,
			KMSMasterKeyID: strPtr("arn:aws:kms:us-east-1:123:key/abc"),
		}},
	}
	result := mapEncryption(rules)
	if result == nil {
		t.Fatal("expected non-nil encryption")
	}
	if result.KmsKeyName != "arn:aws:kms:us-east-1:123:key/abc" {
		t.Errorf("expected KMS key ARN, got %q", result.KmsKeyName)
	}
}

func TestMapEncryption_Nil(t *testing.T) {
	result := mapEncryption(nil)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestMapVersioning(t *testing.T) {
	tests := []struct {
		name   string
		status types.BucketVersioningStatus
		want   bool
	}{
		{"enabled", types.BucketVersioningStatusEnabled, true},
		{"suspended", types.BucketVersioningStatusSuspended, false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapVersioning(tt.status)
			if result.Enabled != tt.want {
				t.Errorf("expected enabled=%v, got %v", tt.want, result.Enabled)
			}
		})
	}
}

func boolPtr(b bool) *bool { return &b }

func TestMapPublicAccessBlock_AllTrue(t *testing.T) {
	cfg := &types.PublicAccessBlockConfiguration{
		BlockPublicAcls:       boolPtr(true),
		BlockPublicPolicy:     boolPtr(true),
		IgnorePublicAcls:      boolPtr(true),
		RestrictPublicBuckets: boolPtr(true),
	}
	result := mapPublicAccessBlock(cfg)
	if result != "Enforced" {
		t.Errorf("expected 'Enforced', got %q", result)
	}
}

func TestMapPublicAccessBlock_Mixed(t *testing.T) {
	cfg := &types.PublicAccessBlockConfiguration{
		BlockPublicAcls:       boolPtr(true),
		BlockPublicPolicy:     boolPtr(false),
		IgnorePublicAcls:      boolPtr(true),
		RestrictPublicBuckets: boolPtr(true),
	}
	result := mapPublicAccessBlock(cfg)
	if result != "Inherited" {
		t.Errorf("expected 'Inherited', got %q", result)
	}
}

func TestMapPublicAccessBlock_Nil(t *testing.T) {
	result := mapPublicAccessBlock(nil)
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestMapLogging_Configured(t *testing.T) {
	logging := &types.LoggingEnabled{
		TargetBucket: strPtr("log-bucket"),
		TargetPrefix: strPtr("access-logs/"),
	}
	result := mapLogging(logging)
	if result == nil {
		t.Fatal("expected non-nil logging")
	}
	if result.LogBucket != "log-bucket" {
		t.Errorf("expected 'log-bucket', got %q", result.LogBucket)
	}
	if result.LogObjectPrefix != "access-logs/" {
		t.Errorf("expected 'access-logs/', got %q", result.LogObjectPrefix)
	}
}

func TestMapLogging_Nil(t *testing.T) {
	result := mapLogging(nil)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestMapRetentionPolicy_Compliance(t *testing.T) {
	days := int32(90)
	config := &types.ObjectLockConfiguration{
		Rule: &types.ObjectLockRule{
			DefaultRetention: &types.DefaultRetention{
				Mode: types.ObjectLockRetentionModeCompliance,
				Days: &days,
			},
		},
	}
	result := mapRetentionPolicy(config)
	if result == nil {
		t.Fatal("expected non-nil retention policy")
	}
	if !result.IsLocked {
		t.Error("expected IsLocked=true for COMPLIANCE mode")
	}
	if result.RetentionPeriod.Hours() != 90*24 {
		t.Errorf("expected 90 days, got %v", result.RetentionPeriod)
	}
}

func TestMapRetentionPolicy_Governance(t *testing.T) {
	days := int32(30)
	config := &types.ObjectLockConfiguration{
		Rule: &types.ObjectLockRule{
			DefaultRetention: &types.DefaultRetention{
				Mode: types.ObjectLockRetentionModeGovernance,
				Days: &days,
			},
		},
	}
	result := mapRetentionPolicy(config)
	if result == nil {
		t.Fatal("expected non-nil retention policy")
	}
	if result.IsLocked {
		t.Error("expected IsLocked=false for GOVERNANCE mode")
	}
}

func TestMapRetentionPolicy_Nil(t *testing.T) {
	result := mapRetentionPolicy(nil)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestParseBucketPolicy_ValidPolicy(t *testing.T) {
	policy := `{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Principal": "*",
			"Action": "s3:GetObject",
			"Resource": "arn:aws:s3:::my-bucket/*"
		}]
	}`
	result, err := parseBucketPolicy(policy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(result))
	}
	stmt := result[0]
	if stmt.Effect != "Allow" {
		t.Errorf("expected 'Allow', got %q", stmt.Effect)
	}
	if len(stmt.Principals) != 1 || stmt.Principals[0] != "*" {
		t.Errorf("expected ['*'], got %v", stmt.Principals)
	}
	if len(stmt.Actions) != 1 || stmt.Actions[0] != "s3:GetObject" {
		t.Errorf("expected ['s3:GetObject'], got %v", stmt.Actions)
	}
}

func TestParseBucketPolicy_WithConditions(t *testing.T) {
	policy := `{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Principal": {"AWS": "arn:aws:iam::123:root"},
			"Action": ["s3:GetObject", "s3:PutObject"],
			"Resource": "arn:aws:s3:::my-bucket/*",
			"Condition": {"StringLike": {"s3:prefix": ["home/", "home/user/*"]}}
		}]
	}`
	result, err := parseBucketPolicy(policy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(result))
	}
	stmt := result[0]
	if len(stmt.Conditions) == 0 {
		t.Fatal("expected conditions, got none")
	}
	prefixes, ok := stmt.Conditions["StringLike"]["s3:prefix"]
	if !ok {
		t.Fatal("expected StringLike/s3:prefix condition")
	}
	if len(prefixes) != 2 {
		t.Errorf("expected 2 prefix values, got %d", len(prefixes))
	}
}

func TestParseBucketPolicy_Empty(t *testing.T) {
	result, err := parseBucketPolicy("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for empty policy, got %v", result)
	}
}

func TestParseBucketPolicy_InvalidJSON(t *testing.T) {
	_, err := parseBucketPolicy("{invalid}")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestExtractFilterPrefix_WithPrefix(t *testing.T) {
	filter := &types.LifecycleRuleFilter{Prefix: strPtr("logs/")}
	result := extractFilterPrefix(filter)
	if result != "logs/" {
		t.Errorf("expected 'logs/', got %q", result)
	}
}

func TestExtractFilterPrefix_WithAnd(t *testing.T) {
	filter := &types.LifecycleRuleFilter{
		And: &types.LifecycleRuleAndOperator{Prefix: strPtr("data/")},
	}
	result := extractFilterPrefix(filter)
	if result != "data/" {
		t.Errorf("expected 'data/', got %q", result)
	}
}

func TestExtractFilterPrefix_Nil(t *testing.T) {
	result := extractFilterPrefix(nil)
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestFlattenStringOrSlice(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  []string
	}{
		{"nil", nil, nil},
		{"string", "*", []string{"*"}},
		{"slice", []interface{}{"a", "b"}, []string{"a", "b"}},
		{"map", map[string]interface{}{"AWS": "arn:aws:iam::123:root"}, []string{"arn:aws:iam::123:root"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenStringOrSlice(tt.input)
			if len(result) != len(tt.want) {
				t.Errorf("expected %d items, got %d: %v", len(tt.want), len(result), result)
				return
			}
			for i, v := range result {
				if v != tt.want[i] {
					t.Errorf("item %d: expected %q, got %q", i, tt.want[i], v)
				}
			}
		})
	}
}

func TestDerefHelpers(t *testing.T) {
	s := "hello"
	if derefString(&s) != "hello" {
		t.Error("derefString failed")
	}
	if derefString(nil) != "" {
		t.Error("derefString nil failed")
	}

	b := true
	if !derefBool(&b) {
		t.Error("derefBool failed")
	}
	if derefBool(nil) {
		t.Error("derefBool nil failed")
	}

	var i32 int32 = 42
	if derefInt32(&i32) != 42 {
		t.Error("derefInt32 failed")
	}
	if derefInt32(nil) != 0 {
		t.Error("derefInt32 nil failed")
	}

	var i64 int64 = 100
	if derefInt64(&i64) != 100 {
		t.Error("derefInt64 failed")
	}
	if derefInt64(nil) != 0 {
		t.Error("derefInt64 nil failed")
	}
}

func TestMapACLGrants_OwnerFallback(t *testing.T) {
	owner := &types.Owner{ID: strPtr("same-id"), DisplayName: strPtr("owner-name")}
	grants := []types.Grant{
		{
			Grantee:    &types.Grantee{Type: types.TypeCanonicalUser, ID: strPtr("same-id")},
			Permission: types.PermissionFullControl,
		},
	}
	result := mapACLGrants(owner, grants)
	if len(result) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(result))
	}
	if result[0].Entity != "owner-name" {
		t.Errorf("expected owner display name fallback, got %q", result[0].Entity)
	}
}

func TestResolveGranteeEntity_NilGrantee(t *testing.T) {
	result := resolveGranteeEntity(nil, nil)
	if result != "unknown" {
		t.Errorf("expected 'unknown', got %q", result)
	}
}

func TestMapLifecycleRules_NoncurrentVersionExpiration(t *testing.T) {
	days := int32(60)
	versions := int32(3)
	rules := []types.LifecycleRule{
		{
			Status: types.ExpirationStatusEnabled,
			NoncurrentVersionExpiration: &types.NoncurrentVersionExpiration{
				NoncurrentDays:          &days,
				NewerNoncurrentVersions: &versions,
			},
		},
	}
	result := mapLifecycleRules(rules)
	if len(result) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(result))
	}
	if result[0].Condition.Age != 60 {
		t.Errorf("expected age 60, got %d", result[0].Condition.Age)
	}
	if result[0].Condition.NumNewerVersions != 3 {
		t.Errorf("expected 3 newer versions, got %d", result[0].Condition.NumNewerVersions)
	}
}

func TestMapRetentionPolicy_WithYears(t *testing.T) {
	years := int32(2)
	config := &types.ObjectLockConfiguration{
		Rule: &types.ObjectLockRule{
			DefaultRetention: &types.DefaultRetention{
				Mode:  types.ObjectLockRetentionModeCompliance,
				Years: &years,
			},
		},
	}
	result := mapRetentionPolicy(config)
	if result == nil {
		t.Fatal("expected non-nil")
	}
	expectedHours := float64(2 * 365 * 24)
	if result.RetentionPeriod.Hours() != expectedHours {
		t.Errorf("expected %v hours, got %v", expectedHours, result.RetentionPeriod.Hours())
	}
}

func TestMapPublicAccessBlock_AllFalse(t *testing.T) {
	cfg := &types.PublicAccessBlockConfiguration{
		BlockPublicAcls:       boolPtr(false),
		BlockPublicPolicy:     boolPtr(false),
		IgnorePublicAcls:      boolPtr(false),
		RestrictPublicBuckets: boolPtr(false),
	}
	result := mapPublicAccessBlock(cfg)
	if result != "Inherited" {
		t.Errorf("expected 'Inherited', got %q", result)
	}
}

func TestParseBucketPolicy_MultiplePrincipals(t *testing.T) {
	policy := `{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Deny",
			"Principal": {"AWS": ["arn:aws:iam::111:root", "arn:aws:iam::222:root"]},
			"Action": "s3:*",
			"Resource": ["arn:aws:s3:::b", "arn:aws:s3:::b/*"]
		}]
	}`
	result, err := parseBucketPolicy(policy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	stmt := result[0]
	if stmt.Effect != "Deny" {
		t.Errorf("expected 'Deny', got %q", stmt.Effect)
	}
	if len(stmt.Principals) != 2 {
		t.Errorf("expected 2 principals, got %d", len(stmt.Principals))
	}
	if len(stmt.Resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(stmt.Resources))
	}
}

func TestFlattenConditions_Nil(t *testing.T) {
	result := flattenConditions(nil)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestFlattenConditions_NonMap(t *testing.T) {
	result := flattenConditions("not a map")
	if result != nil {
		t.Errorf("expected nil for non-map, got %v", result)
	}
}

func TestMapEncryption_EmptyDefault(t *testing.T) {
	rules := []types.ServerSideEncryptionRule{
		{ApplyServerSideEncryptionByDefault: nil},
	}
	result := mapEncryption(rules)
	if result != nil {
		t.Errorf("expected nil for nil default, got %v", result)
	}
}

func TestMapLogging_EmptyFields(t *testing.T) {
	logging := &types.LoggingEnabled{}
	result := mapLogging(logging)
	if result == nil {
		t.Fatal("expected non-nil")
	}
	if result.LogBucket != "" || result.LogObjectPrefix != "" {
		t.Errorf("expected empty fields, got bucket=%q prefix=%q", result.LogBucket, result.LogObjectPrefix)
	}
}

func TestMapRetentionPolicy_NoRule(t *testing.T) {
	config := &types.ObjectLockConfiguration{Rule: nil}
	result := mapRetentionPolicy(config)
	if result != nil {
		t.Errorf("expected nil for no rule, got %v", result)
	}
}

func TestMapRetentionPolicy_NoDefaultRetention(t *testing.T) {
	config := &types.ObjectLockConfiguration{
		Rule: &types.ObjectLockRule{DefaultRetention: nil},
	}
	result := mapRetentionPolicy(config)
	if result != nil {
		t.Errorf("expected nil for no default retention, got %v", result)
	}
}
