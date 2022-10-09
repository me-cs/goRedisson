package goRedisson

import (
	"context"
	"errors"
	"github.com/bits-and-blooms/bitset"
	"strconv"
)

type BitSet interface {
	getSigned(size int32, offset int64) (int64, error)
	setSigned(size int32, offset int64, value int64) (int64, error)
	incrementAndGetSigned(size int32, offset int64, increment int64) (int64, error)
	getUnsigned(size int32, offset int64) (int64, error)
	setUnSigned(size int32, offset int64, value int64) (int64, error)
	incrementAndGetUnSigned(size int32, offset int64, increment int64) (int64, error)
	GetByte(offset int64) (byte, error)
	SetByte(offset int64, value byte) (byte, error)
	incrementAndGetByte(offset int64, increment byte) (byte, error)
	GetShort(offset int64) (int16, error)
	SetShort(offset int64, value int16) (int16, error)
	incrementAndGetShort(offset int64, increment int16) (int16, error)
	GetInt32(offset int32) (int32, error)
	SetInt32(offset int64, value int32) (int32, error)
	incrementAndGetInt32(offset int64, increment int32) (int32, error)
	GetInt64(offset int32) (int64, error)
	SetInt64(offset int64, value int64) (int64, error)
	incrementAndGetInt64(offset int64, increment int64) (int64, error)
}

var (
	_ BitSet = (*goRedissonBitSet)(nil)
)

type goRedissonBitSet struct {
	*goRedissonExpirable
	goRedisson *GoRedisson
}

func NewGoRedissonBitSet(goRedisson *GoRedisson, name string) *goRedissonBitSet {
	return &goRedissonBitSet{
		goRedissonExpirable: NewGoRedissonExpirable(name),
		goRedisson:          goRedisson,
	}
}
func (m *goRedissonBitSet) getSigned(size int32, offset int64) (int64, error) {
	if size > 64 {
		return 0, errors.New("size can't be greater than 64 bits")
	}
	r, err := m.goRedisson.client.Do(context.Background(), "BITFIELD", m.getRawName(), "GET", "i"+strconv.FormatInt(int64(size), 10), offset).Result()
	if err != nil {
		return 0, err
	}
	return transResult2Int64(r)
}

func (m *goRedissonBitSet) setSigned(size int32, offset int64, value int64) (int64, error) {
	if size > 64 {
		return 0, errors.New("size can't be greater than 64 bits")
	}
	r, err := m.goRedisson.client.Do(context.Background(), "BITFIELD", m.getRawName(), "SET", "i"+strconv.FormatInt(int64(size), 10), offset, value).Result()
	if err != nil {
		return 0, err
	}
	return transResult2Int64(r)
}

func (m *goRedissonBitSet) incrementAndGetSigned(size int32, offset int64, increment int64) (int64, error) {
	if size > 64 {
		return 0, errors.New("size can't be greater than 64 bits")
	}
	r, err := m.goRedisson.client.Do(context.Background(), "BITFIELD", m.getRawName(), "INCRBY", "i"+strconv.FormatInt(int64(size), 10), offset, increment).Result()
	if err != nil {
		return 0, err
	}
	return transResult2Int64(r)
}

func transResult2Int64(v interface{}) (int64, error) {
	switch v.(type) {
	case []interface{}:
		return v.([]interface{})[0].(int64), nil
	default:
		return 0, errors.New("can't get data from result")
	}
}

func (m *goRedissonBitSet) getUnsigned(size int32, offset int64) (int64, error) {
	if size > 63 {
		return 0, errors.New("size can't be greater than 63 bits")
	}
	v, err := m.goRedisson.client.Do(context.Background(), "BITFIELD", m.getRawName(), "GET", "u"+strconv.FormatInt(int64(size), 10), offset).Result()
	if err != nil {
		return 0, err
	}
	return transResult2Int64(v)
}

func (m *goRedissonBitSet) setUnSigned(size int32, offset int64, value int64) (int64, error) {
	if size > 63 {
		return 0, errors.New("size can't be greater than 64 bits")
	}
	v, err := m.goRedisson.client.Do(context.Background(), "BITFIELD", m.getRawName(), "SET", "u"+strconv.FormatInt(int64(size), 10), offset, value).Result()
	if err != nil {
		return 0, err
	}
	return transResult2Int64(v)

}

func (m *goRedissonBitSet) incrementAndGetUnSigned(size int32, offset int64, increment int64) (int64, error) {
	if size > 63 {
		return 0, errors.New("size can't be greater than 64 bits")
	}
	v, err := m.goRedisson.client.Do(context.Background(), "BITFIELD", m.getRawName(), "INCRBY", "u"+strconv.FormatInt(int64(size), 10), offset, increment).Result()
	if err != nil {
		return 0, err
	}
	return transResult2Int64(v)
}

func transResult2Byte(v interface{}) (byte, error) {
	switch v.(type) {
	case []interface{}:
		return byte(v.([]interface{})[0].(int64)), nil
	default:
		return 0, errors.New("can't get data from result")
	}
}

func (m *goRedissonBitSet) GetByte(offset int64) (byte, error) {
	r, err := m.goRedisson.client.Do(context.Background(), "BITFIELD", m.getRawName(), "GET", "i8", offset).Result()
	if err != nil {
		return byte(0), err
	}
	return transResult2Byte(r)
}

func (m *goRedissonBitSet) SetByte(offset int64, value byte) (byte, error) {
	r, err := m.goRedisson.client.Do(context.Background(), "BITFIELD", m.getRawName(), "SET", "i8", offset, value).Result()
	if err != nil {
		return byte(0), err
	}
	return transResult2Byte(r)
}

func (m *goRedissonBitSet) incrementAndGetByte(offset int64, increment byte) (byte, error) {
	r, err := m.goRedisson.client.Do(context.Background(), "BITFIELD", m.getRawName(), "INCRBY", "i8", offset, increment).Result()
	if err != nil {
		return byte(0), err
	}
	return transResult2Byte(r)
}

func transResult2Short(v interface{}) (int16, error) {
	switch v.(type) {
	case []interface{}:
		return int16(v.([]interface{})[0].(int64)), nil
	default:
		return 0, errors.New("can't get data from result")
	}
}

func (m *goRedissonBitSet) GetShort(offset int64) (int16, error) {
	r, err := m.goRedisson.client.Do(context.Background(), "BITFIELD", m.getRawName(), "GET", "i16", offset).Result()
	if err != nil {
		return int16(0), err
	}
	return transResult2Short(r)
}

func (m *goRedissonBitSet) SetShort(offset int64, value int16) (int16, error) {
	r, err := m.goRedisson.client.Do(context.Background(), "BITFIELD", m.getRawName(), "SET", "i16", offset, value).Result()
	if err != nil {
		return int16(0), err
	}
	return transResult2Short(r)
}

func (m *goRedissonBitSet) incrementAndGetShort(offset int64, increment int16) (int16, error) {
	r, err := m.goRedisson.client.Do(context.Background(), "BITFIELD", m.getRawName(), "INCRBY", "i16", offset, increment).Result()
	if err != nil {
		return int16(0), err
	}
	return transResult2Short(r)
}

func transResult2Int32(v interface{}) (int32, error) {
	switch v.(type) {
	case []interface{}:
		return int32(v.([]interface{})[0].(int64)), nil
	default:
		return 0, errors.New("can't get data from result")
	}
}

func (m *goRedissonBitSet) GetInt32(offset int32) (int32, error) {
	r, err := m.goRedisson.client.Do(context.Background(), "BITFIELD", m.getRawName(), "GET", "i32", offset).Result()
	if err != nil {
		return int32(0), err
	}
	return transResult2Int32(r)
}

func (m *goRedissonBitSet) SetInt32(offset int64, value int32) (int32, error) {
	r, err := m.goRedisson.client.Do(context.Background(), "BITFIELD", m.getRawName(), "SET", "i32", offset, value).Result()
	if err != nil {
		return int32(0), err
	}
	return transResult2Int32(r)
}

func (m *goRedissonBitSet) incrementAndGetInt32(offset int64, increment int32) (int32, error) {
	r, err := m.goRedisson.client.Do(context.Background(), "BITFIELD", m.getRawName(), "INCRBY", "i32", offset, increment).Result()
	if err != nil {
		return int32(0), err
	}
	return transResult2Int32(r)
}

func (m *goRedissonBitSet) GetInt64(offset int32) (int64, error) {
	r, err := m.goRedisson.client.Do(context.Background(), "BITFIELD", m.getRawName(), "GET", "i64", offset).Result()
	if err != nil {
		return 0, err
	}
	return transResult2Int64(r)
}

func (m *goRedissonBitSet) SetInt64(offset int64, value int64) (int64, error) {
	r, err := m.goRedisson.client.Do(context.Background(), "BITFIELD", m.getRawName(), "SET", "i64", offset, value).Result()
	if err != nil {
		return 0, err
	}
	return transResult2Int64(r)
}

func (m *goRedissonBitSet) incrementAndGetInt64(offset int64, increment int64) (int64, error) {
	r, err := m.goRedisson.client.Do(context.Background(), "BITFIELD", m.getRawName(), "INCRBY", "i64", offset, increment).Result()
	if err != nil {
		return 0, err
	}
	return transResult2Int64(r)
}

func (m *goRedissonBitSet) Set(b bitset.BitSet) error {
	return m.goRedisson.client.Do(context.Background(), "SET", m.getRawName(), b.Bytes()).Err()
}
