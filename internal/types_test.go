package internal

import "testing"

func TestSchema_UnmarshalText(t *testing.T) {
	tests := []struct {
		schemaName  string
		packageName string
	}{
		{schemaName: "hello world", packageName: "hello"},
		{schemaName: "hello_world", packageName: "hello"},
		{schemaName: "12345 Hello World", packageName: "anything"},
		{schemaName: "Hello : World", packageName: "anything"},
		{schemaName: "Hello : World", packageName: ""},
		{schemaName: ": World", packageName: ""},
		{schemaName: "World", packageName: ""},
	}
	for _, tt := range tests {
		input := tt.schemaName + ":" + tt.packageName
		t.Run(input, func(t *testing.T) {
			s := &Schema{}
			if err := s.UnmarshalText([]byte(input)); err != nil {
				t.Errorf("Schema.UnmarshalText() error = %v", err)
				return
			}

			if s.Name != tt.schemaName {
				t.Errorf("wrong schema name, want:%v , got:%v", tt.schemaName, s.Name)
				return
			}

			if s.Package != tt.packageName {
				t.Errorf("wrong package name, want:%v , got:%v", tt.packageName, s.Package)
			}

		})
	}
}
