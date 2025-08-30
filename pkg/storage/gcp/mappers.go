// File: pkg/storage/gcp/mappers.go
package gcp

import (
	"encoding/base64"
	"fmt"
	"synkronus/pkg/storage"

	gcpstorage "cloud.google.com/go/storage"
)

func mapLifecycleRules(rules []gcpstorage.LifecycleRule) []storage.LifecycleRule {
	if len(rules) == 0 {
		return nil
	}
	var result []storage.LifecycleRule
	for _, r := range rules {
		var actionStr string
		// Refine action string for better readability
		if r.Action.StorageClass != "" {
			actionStr = fmt.Sprintf("%s to %s", r.Action.Type, r.Action.StorageClass)
		} else {
			actionStr = r.Action.Type
		}

		result = append(result, storage.LifecycleRule{
			Action: actionStr,
			Condition: storage.LifecycleCondition{
				Age:                 int(r.Condition.AgeInDays),
				CreatedBefore:       r.Condition.CreatedBefore,
				MatchesStorageClass: r.Condition.MatchesStorageClasses,
				NumNewerVersions:    int(r.Condition.NumNewerVersions),
			},
		})
	}
	return result
}

func mapLogging(l *gcpstorage.BucketLogging) *storage.Logging {
	if l == nil {
		return nil
	}
	return &storage.Logging{
		LogBucket:       l.LogBucket,
		LogObjectPrefix: l.LogObjectPrefix,
	}
}

func mapSoftDeletePolicy(sdp *gcpstorage.SoftDeletePolicy) *storage.SoftDeletePolicy {
	if sdp == nil {
		return nil
	}
	return &storage.SoftDeletePolicy{
		RetentionDuration: sdp.RetentionDuration,
	}
}

func mapPublicAccessPrevention(pap gcpstorage.PublicAccessPrevention) string {
	switch pap {
	case gcpstorage.PublicAccessPreventionEnforced:
		return "Enforced"
	case gcpstorage.PublicAccessPreventionInherited:
		return "Inherited"
	default:
		return "Unknown"
	}
}

// Renamed from mapEncryption for clarity, maps bucket-level settings
func mapBucketEncryption(e *gcpstorage.BucketEncryption) *storage.Encryption {
	if e == nil || e.DefaultKMSKeyName == "" {
		// Empty key name implies Google-managed encryption
		return nil
	}
	return &storage.Encryption{
		KmsKeyName: e.DefaultKMSKeyName,
		Algorithm:  "AES256",
	}
}

func mapRetentionPolicy(rp *gcpstorage.RetentionPolicy) *storage.RetentionPolicy {
	if rp == nil {
		return nil
	}
	return &storage.RetentionPolicy{
		RetentionPeriod: rp.RetentionPeriod,
		IsLocked:        rp.IsLocked,
	}
}

// Converts the binary MD5 hash provided by GCP SDK into a standard Base64 encoded string
func formatMD5(hash []byte) string {
	if len(hash) == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(hash)
}

// Converts the uint32 CRC32C checksum provided by GCP SDK into a standard Base64 encoded string
func formatCRC32C(crc32c uint32) string {
	if crc32c == 0 {
		return ""
	}
	// Convert the uint32 to a 4-byte big-endian slice
	b := []byte{
		byte(crc32c >> 24),
		byte(crc32c >> 16),
		byte(crc32c >> 8),
		byte(crc32c),
	}
	return base64.StdEncoding.EncodeToString(b)
}
