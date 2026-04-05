package output

import (
	"fmt"
	"strings"
	"time"

	domainsql "synkronus/internal/domain/sql"
)

// InstanceListView renders a slice of SQL instances as an ASCII table.
type InstanceListView []domainsql.Instance

// RenderTable returns the instance list formatted as an ASCII table.
func (v InstanceListView) RenderTable() string {
	table := NewTable([]string{"INSTANCE NAME", "PROVIDER", "REGION", "VERSION", "TIER", "STATE", "ADDRESS", "CREATED"})

	for _, inst := range v {
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

// InstanceDetailView renders a single SQL instance's full detail as an ASCII table.
type InstanceDetailView struct{ domainsql.Instance }

// RenderTable returns the instance detail formatted as sectioned ASCII tables.
func (v InstanceDetailView) RenderTable() string {
	var sb strings.Builder

	sb.WriteString(FormatHeaderSection("SQL Instance: " + v.Name))
	sb.WriteString("\n\n")

	sb.WriteString(v.renderOverview())
	sb.WriteString(v.renderLabels())

	return sb.String()
}

func (v InstanceDetailView) renderOverview() string {
	var sb strings.Builder

	sb.WriteString(FormatSectionTitle("Overview"))
	sb.WriteString("\n")

	table := NewTable([]string{"Parameter", "Value"})
	table.AddRow([]string{"Provider", string(v.Provider)})
	table.AddRow([]string{"Region", v.Region})
	table.AddRow([]string{"Database Version", v.DatabaseVersion})
	table.AddRow([]string{"Tier", v.Tier})
	table.AddRow([]string{"State", v.State})

	if v.PrimaryAddress != "" {
		table.AddRow([]string{"Primary Address", v.PrimaryAddress})
	}

	if v.StorageSizeGB > 0 {
		table.AddRow([]string{"Storage Size", fmt.Sprintf("%d GB", v.StorageSizeGB)})
	}

	if v.Project != "" {
		table.AddRow([]string{"Project", v.Project})
	}

	if v.ConnectionName != "" {
		table.AddRow([]string{"Connection Name", v.ConnectionName})
	}

	if !v.CreatedAt.IsZero() {
		table.AddRow([]string{"Created On", v.CreatedAt.Format(time.RFC1123)})
	}

	sb.WriteString(table.String())
	sb.WriteString("\n\n")

	return sb.String()
}

func (v InstanceDetailView) renderLabels() string {
	return renderLabelsSection(v.Labels)
}
