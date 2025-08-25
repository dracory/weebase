package weebase

import "strings"

// splitCSV splits by comma and trims spaces; ignores empty items.
func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

// sanitizeIdent allows only letters, digits, underscore and dot.
func sanitizeIdent(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '.' {
			continue
		}
		return false
	}
	return true
}

// normalizeDriver normalizes common driver aliases to canonical names.
func normalizeDriver(d string) string {
	switch strings.ToLower(d) {
	case "pg", "postgresql":
		return "postgres"
	case "mariadb":
		return "mysql"
	case "sqlite3":
		return "sqlite"
	case "mssql":
		return "sqlserver"
	default:
		return strings.ToLower(d)
	}
}

// quoteIdent safely quotes a possibly schema-qualified identifier for the given driver.
// It assumes parts have already passed sanitizeIdent (letters/digits/underscore/dot).
// This is a best-effort helper for building simple SQL strings.
func quoteIdent(driver, ident string) string {
	d := normalizeDriver(driver)
	// split by '.' to support schema.table
	parts := strings.Split(ident, ".")
	for i, p := range parts {
		switch d {
		case "postgres":
			// escape double quotes by doubling them
			p = strings.ReplaceAll(p, "\"", "\"\"")
			parts[i] = "\"" + p + "\""
		case "mysql", "sqlite":
			// escape backticks by doubling them
			p = strings.ReplaceAll(p, "`", "``")
			parts[i] = "`" + p + "`"
		case "sqlserver":
			// escape ']' by doubling it
			p = strings.ReplaceAll(p, "]", "]]")
			parts[i] = "[" + p + "]"
		default:
			parts[i] = p
		}
	}
	return strings.Join(parts, ".")
}
