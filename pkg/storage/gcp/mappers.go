// File: pkg/storage/gcp/mappers.go
package gcp

import (
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
