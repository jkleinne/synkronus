package aws

import (
	"encoding/json"
	"fmt"
	"synkronus/internal/domain/storage"
	"synkronus/internal/provider/storage/shared"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func mapTags(tags []types.Tag) map[string]string {
	if len(tags) == 0 {
		return nil
	}
	result := make(map[string]string, len(tags))
	for _, t := range tags {
		result[derefString(t.Key)] = derefString(t.Value)
	}
	return result
}

func mapACLGrants(owner *types.Owner, grants []types.Grant) []storage.ACLRule {
	if len(grants) == 0 {
		return nil
	}
	var rules []storage.ACLRule
	for _, g := range grants {
		entity := resolveGranteeEntity(g.Grantee, owner)
		rules = append(rules, storage.ACLRule{
			Entity: entity,
			Role:   string(g.Permission),
		})
	}
	return rules
}

func resolveGranteeEntity(grantee *types.Grantee, owner *types.Owner) string {
	if grantee == nil {
		return "unknown"
	}
	switch grantee.Type {
	case types.TypeCanonicalUser:
		displayName := derefString(grantee.DisplayName)
		if displayName != "" {
			return displayName
		}
		id := derefString(grantee.ID)
		if owner != nil && id == derefString(owner.ID) {
			return derefString(owner.DisplayName)
		}
		return id
	case types.TypeGroup:
		return derefString(grantee.URI)
	case types.TypeAmazonCustomerByEmail:
		return derefString(grantee.EmailAddress)
	default:
		return "unknown"
	}
}

func mapLifecycleRules(rules []types.LifecycleRule) []storage.LifecycleRule {
	if len(rules) == 0 {
		return nil
	}
	var result []storage.LifecycleRule

	for _, r := range rules {
		prefix := extractFilterPrefix(r.Filter)
		// Skip disabled rules
		if r.Status != types.ExpirationStatusEnabled {
			continue
		}

		// Map expiration → "Delete" action
		if r.Expiration != nil && r.Expiration.Days != nil {
			result = append(result, storage.LifecycleRule{
				Action: "Delete",
				Condition: storage.LifecycleCondition{
					Age:    int(*r.Expiration.Days),
					Prefix: prefix,
				},
			})
		}

		// Map transitions → "Transition to <CLASS>" action
		for _, t := range r.Transitions {
			action := fmt.Sprintf("Transition to %s", t.StorageClass)
			condition := storage.LifecycleCondition{Prefix: prefix}
			if t.Days != nil {
				condition.Age = int(*t.Days)
			}
			result = append(result, storage.LifecycleRule{
				Action:    action,
				Condition: condition,
			})
		}

		// Map noncurrent version expiration
		if r.NoncurrentVersionExpiration != nil && r.NoncurrentVersionExpiration.NoncurrentDays != nil {
			result = append(result, storage.LifecycleRule{
				Action: "Delete",
				Condition: storage.LifecycleCondition{
					Age:              int(*r.NoncurrentVersionExpiration.NoncurrentDays),
					NumNewerVersions: int(derefInt32(r.NoncurrentVersionExpiration.NewerNoncurrentVersions)),
					Prefix:           prefix,
				},
			})
		}
	}

	return result
}

func extractFilterPrefix(filter *types.LifecycleRuleFilter) string {
	if filter == nil {
		return ""
	}
	if filter.Prefix != nil {
		return *filter.Prefix
	}
	if filter.And != nil && filter.And.Prefix != nil {
		return *filter.And.Prefix
	}
	return ""
}

func mapEncryption(rules []types.ServerSideEncryptionRule) *storage.Encryption {
	if len(rules) == 0 {
		return nil
	}
	rule := rules[0]
	if rule.ApplyServerSideEncryptionByDefault == nil {
		return nil
	}
	enc := rule.ApplyServerSideEncryptionByDefault
	result := &storage.Encryption{
		Algorithm: string(enc.SSEAlgorithm),
	}
	if kmsKey := derefString(enc.KMSMasterKeyID); kmsKey != "" {
		result.KmsKeyName = kmsKey
	}
	return result
}

func mapVersioning(status types.BucketVersioningStatus) *storage.Versioning {
	return &storage.Versioning{
		Enabled: status == types.BucketVersioningStatusEnabled,
	}
}

func mapPublicAccessBlock(cfg *types.PublicAccessBlockConfiguration) string {
	if cfg == nil {
		return ""
	}
	if derefBool(cfg.BlockPublicAcls) && derefBool(cfg.BlockPublicPolicy) &&
		derefBool(cfg.IgnorePublicAcls) && derefBool(cfg.RestrictPublicBuckets) {
		return shared.PublicAccessEnforced
	}
	return shared.PublicAccessInherited
}

func derefBool(p *bool) bool {
	if p == nil {
		return false
	}
	return *p
}

func mapLogging(logging *types.LoggingEnabled) *storage.Logging {
	if logging == nil {
		return nil
	}
	return &storage.Logging{
		LogBucket:       derefString(logging.TargetBucket),
		LogObjectPrefix: derefString(logging.TargetPrefix),
	}
}

func mapRetentionPolicy(config *types.ObjectLockConfiguration) *storage.RetentionPolicy {
	if config == nil || config.Rule == nil || config.Rule.DefaultRetention == nil {
		return nil
	}
	ret := config.Rule.DefaultRetention
	var period time.Duration
	if ret.Days != nil {
		period = time.Duration(*ret.Days) * 24 * time.Hour
	} else if ret.Years != nil {
		period = time.Duration(*ret.Years) * 365 * 24 * time.Hour
	}
	return &storage.RetentionPolicy{
		RetentionPeriod: period,
		IsLocked:        ret.Mode == types.ObjectLockRetentionModeCompliance,
	}
}

// policyDocument represents the JSON structure of an S3 bucket policy.
type policyDocument struct {
	Version   string            `json:"Version"`
	Statement []policyStatement `json:"Statement"`
}

type policyStatement struct {
	Effect    string `json:"Effect"`
	Principal any    `json:"Principal"` // Can be string ("*") or map
	Action    any    `json:"Action"`    // Can be string or []string
	Resource  any    `json:"Resource"`  // Can be string or []string
	Condition any    `json:"Condition,omitempty"`
}

func parseBucketPolicy(policyJSON string) ([]storage.PolicyStatement, error) {
	if policyJSON == "" {
		return nil, nil
	}

	var doc policyDocument
	if err := json.Unmarshal([]byte(policyJSON), &doc); err != nil {
		return nil, fmt.Errorf("failed to parse bucket policy: %w", err)
	}

	var statements []storage.PolicyStatement
	for _, s := range doc.Statement {
		stmt := storage.PolicyStatement{
			Effect:     s.Effect,
			Principals: flattenStringOrSlice(s.Principal),
			Actions:    flattenStringOrSlice(s.Action),
			Resources:  flattenStringOrSlice(s.Resource),
			Conditions: flattenConditions(s.Condition),
		}
		statements = append(statements, stmt)
	}
	return statements, nil
}

// flattenStringOrSlice handles JSON fields that can be a string, a []string, or a map with a key like "AWS".
func flattenStringOrSlice(v any) []string {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case string:
		return []string{val}
	case []any:
		result := make([]string, 0, len(val))
		for _, item := range val {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	case map[string]any:
		// Handle {"AWS": "arn:..."} or {"AWS": ["arn:...", "arn:..."]}
		var result []string
		for _, sub := range val {
			result = append(result, flattenStringOrSlice(sub)...)
		}
		return result
	default:
		return nil
	}
}

// flattenConditions parses the Condition block of an IAM policy statement.
func flattenConditions(v any) map[string]map[string][]string {
	if v == nil {
		return nil
	}
	raw, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	result := make(map[string]map[string][]string, len(raw))
	for operator, keysRaw := range raw {
		keys, ok := keysRaw.(map[string]any)
		if !ok {
			continue
		}
		inner := make(map[string][]string, len(keys))
		for key, valRaw := range keys {
			inner[key] = flattenStringOrSlice(valRaw)
		}
		result[operator] = inner
	}
	return result
}

func derefInt32(p *int32) int32 {
	if p == nil {
		return 0
	}
	return *p
}
