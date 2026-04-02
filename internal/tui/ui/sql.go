package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	domainsql "synkronus/internal/domain/sql"
)

// RenderInstanceList renders the SQL instance table with cursor and scroll.
// Returns a centered "No instances found" message when instances is nil or empty.
func RenderInstanceList(instances []domainsql.Instance, cursor, offset, termWidth int) string {
	if len(instances) == 0 {
		msg := TextDimStyle.Render("No instances found")
		return lipgloss.PlaceHorizontal(termWidth, lipgloss.Center, msg)
	}

	headers := []string{"Name", "Provider", "Region", "Version", "Tier", "State", "Address", "Created"}
	rows := make([][]string, len(instances))
	for i, inst := range instances {
		createdFormatted := ""
		if !inst.CreatedAt.IsZero() {
			createdFormatted = inst.CreatedAt.Format("2006-01-02")
		}
		rows[i] = []string{
			inst.Name,
			string(inst.Provider),
			inst.Region,
			inst.DatabaseVersion,
			inst.Tier,
			inst.State,
			inst.PrimaryAddress,
			createdFormatted,
		}
	}

	return RenderTable(headers, rows, cursor, offset, termWidth)
}

// RenderInstanceDetail renders instance metadata in key-value sections.
// Includes an Overview section with core fields and an optional Labels section.
func RenderInstanceDetail(instance domainsql.Instance, termWidth int) string {
	overviewEntries := []KeyValue{
		{Key: "Name", Value: instance.Name},
		{Key: "Provider", Value: string(instance.Provider), Style: ValueProvider},
		{Key: "Region", Value: instance.Region},
		{Key: "Database Version", Value: instance.DatabaseVersion},
		{Key: "Tier", Value: instance.Tier},
		{Key: "State", Value: instance.State},
		{Key: "Primary Address", Value: instance.PrimaryAddress},
		{Key: "Connection Name", Value: instance.ConnectionName},
		{Key: "Storage Size", Value: fmt.Sprintf("%d GB", instance.StorageSizeGB)},
		{Key: "Project", Value: instance.Project},
	}

	if !instance.CreatedAt.IsZero() {
		overviewEntries = append(overviewEntries, KeyValue{
			Key:   "Created",
			Value: instance.CreatedAt.Format("2006-01-02"),
		})
	}

	sections := []KeyValueSection{
		{Title: "Overview", Entries: overviewEntries},
	}

	if len(instance.Labels) > 0 {
		labelEntries := make([]KeyValue, 0, len(instance.Labels))
		for k, v := range instance.Labels {
			labelEntries = append(labelEntries, KeyValue{Key: k, Value: v})
		}
		sections = append(sections, KeyValueSection{Title: "Labels", Entries: labelEntries})
	}

	return RenderKeyValueGrid(sections, termWidth)
}
