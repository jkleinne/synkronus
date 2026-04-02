// File: pkg/formatter/sql_formatter.go
package formatter

import (
	"fmt"
	"strings"
	"synkronus/internal/domain/sql"
	"time"
)

type SqlFormatter struct{}

func NewSqlFormatter() *SqlFormatter {
	return &SqlFormatter{}
}

// FormatInstanceList formats a list of SQL instances as a table
func (f *SqlFormatter) FormatInstanceList(instances []sql.Instance) string {
	table := NewTable([]string{"INSTANCE NAME", "PROVIDER", "REGION", "VERSION", "TIER", "STATE", "ADDRESS", "CREATED"})

	for _, inst := range instances {
		createdFormatted := ""
		if !inst.CreatedAt.IsZero() {
			createdFormatted = inst.CreatedAt.Format("2006-01-02")
		}

		table.AddRow([]string{
			inst.Name,
			string(inst.Provider),
			inst.Region,
			inst.DatabaseVersion,
			inst.Tier,
			inst.State,
			inst.PrimaryAddress,
			createdFormatted,
		})
	}

	return table.String()
}

// FormatInstanceDetails formats detailed information about a single SQL instance
func (f *SqlFormatter) FormatInstanceDetails(instance sql.Instance) string {
	var sb strings.Builder

	sb.WriteString(FormatHeaderSection("SQL Instance: " + instance.Name))
	sb.WriteString("\n\n")

	sb.WriteString(f.formatInstanceOverviewSection(instance))
	sb.WriteString(f.formatInstanceLabelsSection(instance))

	return sb.String()
}

func (f *SqlFormatter) formatInstanceOverviewSection(instance sql.Instance) string {
	var sb strings.Builder

	sb.WriteString(FormatSectionTitle("Overview"))
	sb.WriteString("\n")

	overviewTable := NewTable([]string{"Parameter", "Value"})
	overviewTable.AddRow([]string{"Provider", string(instance.Provider)})
	overviewTable.AddRow([]string{"Region", instance.Region})
	overviewTable.AddRow([]string{"Database Version", instance.DatabaseVersion})
	overviewTable.AddRow([]string{"Tier", instance.Tier})
	overviewTable.AddRow([]string{"State", instance.State})

	if instance.PrimaryAddress != "" {
		overviewTable.AddRow([]string{"Primary Address", instance.PrimaryAddress})
	}

	if instance.StorageSizeGB > 0 {
		overviewTable.AddRow([]string{"Storage Size", fmt.Sprintf("%d GB", instance.StorageSizeGB)})
	}

	if instance.Project != "" {
		overviewTable.AddRow([]string{"Project", instance.Project})
	}

	if instance.ConnectionName != "" {
		overviewTable.AddRow([]string{"Connection Name", instance.ConnectionName})
	}

	if !instance.CreatedAt.IsZero() {
		overviewTable.AddRow([]string{"Created On", instance.CreatedAt.Format(time.RFC1123)})
	}

	sb.WriteString(overviewTable.String())
	sb.WriteString("\n\n")

	return sb.String()
}

func (f *SqlFormatter) formatInstanceLabelsSection(instance sql.Instance) string {
	if len(instance.Labels) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(FormatSectionTitle("Labels"))
	sb.WriteString("\n")
	labelsTable := NewTable([]string{"Key", "Value"})
	for k, v := range instance.Labels {
		labelsTable.AddRow([]string{k, v})
	}
	sb.WriteString(labelsTable.String())
	sb.WriteString("\n\n")

	return sb.String()
}
