package workspace

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// SheetSize holds the data extent of a sheet.
type SheetSize struct {
	RowCount  int
	ColCount  int
	DataRange string // e.g. "A1:H120", ready to pass directly to ReadSheet
}

// ListSheets returns the names of all sheets in the spreadsheet.
func (c *Client) ListSheets(ctx context.Context, spreadsheetID string) ([]string, error) {
	resp, err := c.sheetsService.Spreadsheets.Get(spreadsheetID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get spreadsheet %s: %w", spreadsheetID, err)
	}

	names := make([]string, len(resp.Sheets))
	for i, sheet := range resp.Sheets {
		names[i] = sheet.Properties.Title
	}
	return names, nil
}

// GetSheetSize returns the extent of actual data in a sheet by fetching all values
// and finding the last populated row and column.
func (c *Client) GetSheetSize(ctx context.Context, spreadsheetID, sheetName string) (SheetSize, error) {
	resp, err := c.sheetsService.Spreadsheets.Values.
		Get(spreadsheetID, quotedSheetName(sheetName)).
		ValueRenderOption("UNFORMATTED_VALUE").
		Context(ctx).
		Do()
	if err != nil {
		return SheetSize{}, fmt.Errorf("failed to get sheet size for %q: %w", sheetName, err)
	}

	rowCount := len(resp.Values)
	if rowCount == 0 {
		return SheetSize{}, nil
	}

	colCount := 0
	for _, row := range resp.Values {
		if len(row) > colCount {
			colCount = len(row)
		}
	}

	return SheetSize{
		RowCount:  rowCount,
		ColCount:  colCount,
		DataRange: "A1:" + colToLetter(colCount) + strconv.Itoa(rowCount),
	}, nil
}

// ReadSheet reads cells from a sheet and returns them as a 2D array.
// If rangeA1 is empty, the full data extent is fetched automatically via GetSheetSize.
func (c *Client) ReadSheet(ctx context.Context, spreadsheetID, sheetName, rangeA1 string) ([][]any, error) {
	if rangeA1 == "" {
		size, err := c.GetSheetSize(ctx, spreadsheetID, sheetName)
		if err != nil {
			return nil, err
		}
		if size.DataRange == "" {
			return [][]any{}, nil
		}
		rangeA1 = size.DataRange
	}

	resp, err := c.sheetsService.Spreadsheets.Values.
		Get(spreadsheetID, quotedSheetName(sheetName)+"!"+rangeA1).
		ValueRenderOption("UNFORMATTED_VALUE").
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to read sheet %q range %s: %w", sheetName, rangeA1, err)
	}

	if resp.Values == nil {
		return [][]any{}, nil
	}
	return resp.Values, nil
}

// colToLetter converts a 1-indexed column number to its A1 notation letter(s).
// e.g. 1→"A", 26→"Z", 27→"AA", 53→"BA".
func colToLetter(n int) string {
	result := ""
	for n > 0 {
		n--
		result = string(rune('A'+n%26)) + result
		n /= 26
	}
	return result
}

// quotedSheetName wraps a sheet name in single quotes for safe use in A1 range notation,
// escaping any single quotes within the name as ”.
func quotedSheetName(name string) string {
	return "'" + strings.ReplaceAll(name, "'", "''") + "'"
}
