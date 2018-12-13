func UnmarshalDatetime(v interface{}) (time.Time, error) {
	if str, ok := v.(string); ok {
		layout := "2006-01-02 15:04:05"
		return time.Parse(layout, str)
	}
	return time.Time{}, errors.New("time should be a unix timestamp")
}

func MarshalDatetime(t time.Time) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, t.Format("2006-01-02 15:04:05"))
	})
}

func MarshalDatetime(t time.Time) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, t.Format("2006-01-02 15:04:05"))
	})
}

func UnmarshalNullTime(v interface{}) (mysql.NullTime, error) {
	nt := mysql.NullTime{}
	if str, ok := v.(string); ok {
		layout := "2006-01-02 15:04:05"
		t, err := time.Parse(layout, str)
		if err == nil {
			nt.Time = t
			nt.Valid = true
		}
		return nt, err
	}
	return nt, errors.New("time should be a unix timestamp")
}

func UnmarshalIntBool(v interface{}) (int8, error) {
	if value, ok := v.(bool); ok {
		if value {
			return 1, nil
		}
		return 0, nil
	}
	return 0, errors.New("value is not boolean")
}

func MarshalIntBool(v int8) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		if v == 1 {
			io.WriteString(w, "true")
		} else {
			io.WriteString(w, "false")
		}
	})
}