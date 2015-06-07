// Tideland Go Library - Redis Client - Result Set
//
// Copyright (C) 2009-2015 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package redis

//--------------------
// IMPORTS
//--------------------

import (
	"fmt"
	"strings"

	"github.com/tideland/golib/errors"
)

//--------------------
// RESULT SET
//--------------------

// ResultSet contains a number of values or nested result sets.
type ResultSet struct {
	parent *ResultSet
	items  []interface{}
	length int
}

// newResultSet creates a new result set.
func newResultSet() *ResultSet {
	return &ResultSet{nil, []interface{}{}, 1}
}

// append adds a value/result set to the result set. It panics if it's
// neither a value, even as a byte slice, nor an array.
func (rs *ResultSet) append(item interface{}) {
	switch i := item.(type) {
	case Value, *ResultSet:
		rs.items = append(rs.items, i)
	case []byte:
		rs.items = append(rs.items, Value(i))
	case ResultSet:
		rs.items = append(rs.items, &i)
	default:
		panic("illegal result set item type")
	}
}

// allReceived answers with true if all expected items are received.
func (rs *ResultSet) allReceived() bool {
	return len(rs.items) >= rs.length
}

// nextResultSet returns the parent stack upwards as long as all expected
// items are received.
func (rs *ResultSet) nextResultSet() *ResultSet {
	if !rs.allReceived() {
		return rs
	}
	if rs.parent == nil {
		return nil
	}
	return rs.parent.nextResultSet()
}

// Len returns the number of items in the result set.
func (rs *ResultSet) Len() int {
	return len(rs.items)
}

// ValueAt returns the value at index.
func (rs *ResultSet) ValueAt(index int) (Value, error) {
	if len(rs.items) < index+1 {
		return nil, errors.New(ErrIllegalItemIndex, errorMessages, index, len(rs.items))
	}
	value, ok := rs.items[index].(Value)
	if !ok {
		return nil, errors.New(ErrIllegalItemType, errorMessages, index, "value")
	}
	return value, nil
}

// BoolAt returns the value at index as bool. This is a convenience
// method as the bool is needed very often.
func (rs *ResultSet) BoolAt(index int) (bool, error) {
	value, err := rs.ValueAt(index)
	if err != nil {
		return false, err
	}
	return value.Bool()
}

// IntAt returns the value at index as int. This is a convenience
// method as the integer is needed very often.
func (rs *ResultSet) IntAt(index int) (int, error) {
	value, err := rs.ValueAt(index)
	if err != nil {
		return 0, err
	}
	return value.Int()
}

// StringAt returns the value at index as string. This is a convenience
// method as the string is needed very often.
func (rs *ResultSet) StringAt(index int) (string, error) {
	value, err := rs.ValueAt(index)
	if err != nil {
		return "", err
	}
	return value.String(), nil
}

// ResultSetAt returns the nested result set at index.
func (rs *ResultSet) ResultSetAt(index int) (*ResultSet, error) {
	if len(rs.items) < index-1 {
		return nil, errors.New(ErrIllegalItemIndex, errorMessages, index, len(rs.items))
	}
	resultSet, ok := rs.items[index].(*ResultSet)
	if !ok {
		return nil, errors.New(ErrIllegalItemType, errorMessages, index, "result set")
	}
	return resultSet, nil
}

// Values returnes a flattened list of all values.
func (rs *ResultSet) Values() Values {
	values := []Value{}
	for _, item := range rs.items {
		switch i := item.(type) {
		case Value:
			values = append(values, i)
		case *ResultSet:
			values = append(values, i.Values()...)
		}
	}
	return values
}

// KeyValues returns the alternating values as key/value slice.
func (rs *ResultSet) KeyValues() (KeyValues, error) {
	kvs := KeyValues{}
	key := ""
	for index, item := range rs.items {
		value, ok := item.(Value)
		if !ok {
			return nil, errors.New(ErrIllegalItemType, errorMessages, index, "value")
		}
		if index%2 == 0 {
			key = value.String()
		} else {
			kvs = append(kvs, KeyValue{key, value})
		}
	}
	return kvs, nil
}

// ScoredValues returns the alternating values as scored values slice. If
// withscores is false the result set contains no scores and so they are
// set to 0.0 in the returned scored values.
func (rs *ResultSet) ScoredValues(withscores bool) (ScoredValues, error) {
	svs := ScoredValues{}
	sv := ScoredValue{}
	for index, item := range rs.items {
		value, ok := item.(Value)
		if !ok {
			return nil, errors.New(ErrIllegalItemType, errorMessages, index, "value")
		}
		if withscores {
			// With scores, so alternating values and scores.
			if index%2 == 0 {
				sv.Value = value
			} else {
				score, err := value.Float64()
				if err != nil {
					return nil, err
				}
				sv.Score = score
				svs = append(svs, sv)
				sv = ScoredValue{}
			}
		} else {
			// No scores, only values.
			sv.Value = value
			svs = append(svs, sv)
			sv = ScoredValue{}
		}
	}
	return svs, nil
}

// Hash returns the values of the result set as hash.
func (rs *ResultSet) Hash() (Hash, error) {
	hash := make(Hash)
	key := ""
	for index, item := range rs.items {
		value, ok := item.(Value)
		if !ok {
			return nil, errors.New(ErrIllegalItemType, errorMessages, index, "value")
		}
		if index%2 == 0 {
			key = value.String()
		} else {
			hash.Set(key, value.Bytes())
		}
	}
	return hash, nil
}

// Scanned returns the cursor and the keys or values of a
// scan operation.
func (rs *ResultSet) Scanned() (int, *ResultSet, error) {
	cursor, err := rs.IntAt(0)
	if err != nil {
		return 0, nil, err
	}
	result, err := rs.ResultSetAt(1)
	return cursor, result, err
}

// Strings returns all values/arrays of the array as a slice of strings.
func (rs *ResultSet) Strings() []string {
	ss := make([]string, len(rs.items))
	for index, item := range rs.items {
		s, ok := item.(fmt.Stringer)
		if !ok {
			// Must not happen!
			panic("illegal type in array")
		}
		ss[index] = s.String()
	}
	return ss
}

// String returns the result set in a human readable form.
func (rs *ResultSet) String() string {
	out := "RESULT SET ("
	ss := rs.Strings()
	return out + strings.Join(ss, " / ") + ")"
}

// EOF
