package internal

import (
	"errors"
	"strings"

	"github.com/kenshaw/snaker"
)

// FkMode represents the different foreign key naming modes.
type FkMode int

const (
	// FkModeSmart is the default FkMode.
	//
	// When there are no naming conflicts, FkModeSmart behaves the same
	// FkModeParent, otherwise it behaves the same as FkModeField.
	FkModeSmart FkMode = iota

	// FkModeParent causes a foreign key field to be named in the form of
	// "<type>.<ParentName>".
	//
	// For example, if you have an `authors` and `books` tables, then the
	// foreign key func will be Book.Author.
	FkModeParent

	// FkModeField causes a foreign key field to be named in the form of
	// "<type>.<ParentName>By<Field>".
	//
	// For example, if you have an `authors` and `books` tables, then the
	// foreign key func will be Book.AuthorByAuthorID.
	FkModeField

	// FkModeKey causes a foreign key field to be named in the form of
	// "<type>.<ParentName>By<ForeignKeyName>".
	//
	// For example, if you have an `authors` and `books` tables with a foreign
	// key name of 'fk_123', then the foreign key func will be
	// Book.AuthorByFk123.
	FkModeKey
)

// UnmarshalText unmarshals FkMode from text.
func (f *FkMode) UnmarshalText(text []byte) error {
	switch strings.ToLower(string(text)) {
	case "smart", "default":
		*f = FkModeSmart
	case "parent":
		*f = FkModeParent
	case "field":
		*f = FkModeField
	case "key":
		*f = FkModeKey

	default:
		return errors.New("invalid FkMode")
	}

	return nil
}

// String satisfies the Stringer interface.
func (f FkMode) String() string {
	switch f {
	case FkModeSmart:
		return "smart"
	case FkModeParent:
		return "parent"
	case FkModeField:
		return "field"
	case FkModeKey:
		return "key"
	}

	return "unknown"
}

// fkName returns the name for the foreign key.
func fkName(mode FkMode, fkMap map[string]*ForeignKey, fk *ForeignKey) string {
	switch mode {
	case FkModeParent:
		return fk.RefType.Name
	case FkModeField:
		return fk.RefType.Name + "By" + fk.Field.Name
	case FkModeKey:
		return fk.RefType.Name + "By" + snaker.SnakeToCamelIdentifier(fk.ForeignKey.ForeignKeyName)
	}

	// mode is FkModeSmart
	// inspect all foreign keys and use FkModeField if conflict found
	for _, f := range fkMap {
		if fk != f && fk.Type.Name == f.Type.Name && fk.RefType.Name == f.RefType.Name {
			return fkName(FkModeField, fkMap, fk)
		}
	}

	// no conflict, so use FkModeParent
	return fkName(FkModeParent, fkMap, fk)
}

// ForeignKeyName returns the foreign key name for the passed type.
func (a *ArgType) ForeignKeyName(fkMap map[string]*ForeignKey, fk *ForeignKey) string {
	return fkName(*a.ForeignKeyMode, fkMap, fk)
}
