package binarytable

// IndexOfInt searches an INT column and returns the first matching row, or -1.
func (t *Table) IndexOfInt(value int32, column, start int) (int, error) {
	if start < 0 {
		start = 0
	}
	for row := start; row < t.Header.MaxRows(); row++ {
		v, err := t.GetInt(row, column)
		if err != nil {
			return -1, err
		}
		if v == value {
			return row, nil
		}
	}
	return -1, nil
}

// IndexOfIntByName searches a named INT column.
func (t *Table) IndexOfIntByName(value int32, columnName string, start int) (int, error) {
	info, err := t.columnByName(columnName)
	if err != nil {
		return -1, err
	}
	return t.IndexOfInt(value, info.Index, start)
}

// IndexOfLong searches a LONG column.
func (t *Table) IndexOfLong(value int64, column, start int) (int, error) {
	if start < 0 {
		start = 0
	}
	for row := start; row < t.Header.MaxRows(); row++ {
		v, err := t.GetLong(row, column)
		if err != nil {
			return -1, err
		}
		if v == value {
			return row, nil
		}
	}
	return -1, nil
}

// IndexOfLongByName searches a named LONG column.
func (t *Table) IndexOfLongByName(value int64, columnName string, start int) (int, error) {
	info, err := t.columnByName(columnName)
	if err != nil {
		return -1, err
	}
	return t.IndexOfLong(value, info.Index, start)
}

// IndexOfDouble searches a DOUBLE column.
func (t *Table) IndexOfDouble(value float64, column, start int) (int, error) {
	if start < 0 {
		start = 0
	}
	for row := start; row < t.Header.MaxRows(); row++ {
		v, err := t.GetDouble(row, column)
		if err != nil {
			return -1, err
		}
		if v == value {
			return row, nil
		}
	}
	return -1, nil
}

// IndexOfDoubleByName searches a named DOUBLE column.
func (t *Table) IndexOfDoubleByName(value float64, columnName string, start int) (int, error) {
	info, err := t.columnByName(columnName)
	if err != nil {
		return -1, err
	}
	return t.IndexOfDouble(value, info.Index, start)
}

// IndexOfString searches a STRING column.
func (t *Table) IndexOfString(value string, column, start int) (int, error) {
	if start < 0 {
		start = 0
	}
	for row := start; row < t.Header.MaxRows(); row++ {
		v, err := t.GetString(row, column)
		if err != nil {
			return -1, err
		}
		if v == value {
			return row, nil
		}
	}
	return -1, nil
}

// IndexOfStringByName searches a named STRING column.
func (t *Table) IndexOfStringByName(value string, columnName string, start int) (int, error) {
	info, err := t.columnByName(columnName)
	if err != nil {
		return -1, err
	}
	return t.IndexOfString(value, info.Index, start)
}
