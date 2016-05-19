// Tideland Go Library - Redis Client - Unit Tests - Commands
//
// Copyright (C) 2009-2016 Frank Mueller / Oldenburg / Germany
//
// All rights reserved. Use of this source code is governed
// by the new BSD license.

package redis_test

//--------------------
// IMPORTS
//--------------------

import (
	"testing"
	"time"

	"github.com/tideland/golib/audit"
	"github.com/tideland/golib/errors"
	"github.com/tideland/golib/redis"
)

//--------------------
// TESTS
//--------------------

func TestSimpleKeyOperations(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	conn, restore := connectDatabase(assert)
	defer restore()

	ok, err := conn.DoOK("set", "sko:a", 1)
	assert.Nil(err)
	assert.True(ok)

	skoA, err := conn.DoInt("get", "sko:a")
	assert.Nil(err)
	assert.Equal(skoA, 1)

	exists, err := conn.DoBool("exists", "sko:a")
	assert.Nil(err)
	assert.True(exists)

	conn.Do("set", "sko:b", 2)
	conn.Do("set", "sko:c", 3)
	conn.Do("set", "sko:d", 4)
	conn.Do("set", "sko:e", 5)

	dbSize, err := conn.DoInt("dbsize")
	assert.Nil(err)
	assert.Equal(dbSize, 5)

	keys, err := conn.DoStrings("keys", "sko:*")
	assert.Nil(err)
	assert.Length(keys, 5)

	deleted, err := conn.DoInt("del", "sko:c", "sko:d", "sko:z")
	assert.Nil(err)
	assert.Equal(deleted, 2)

	keys, err = conn.DoStrings("keys", "sko:*")
	assert.Nil(err)
	assert.Length(keys, 3)

	h := redis.NewFilledHash(map[string]interface{}{
		"sko:f": 6,
		"sko:g": 7,
		"sko:h": 8,
	})
	conn.Do("mset", h)

	keys, err = conn.DoStrings("keys", "sko:*")
	assert.Nil(err)
	assert.Length(keys, 6)

	ssIn := []string{"do", "re", "mi", "fa", "sol", "la", "ti"}
	conn.Do("set", "sko:zz", ssIn)
	vOut, err := conn.DoValue("get", "sko:zz")
	assert.Nil(err)
	ssOut := vOut.StringSlice()
	assert.Equal(ssOut, ssIn)
}

func TestScan(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	conn, restore := connectDatabase(assert)
	defer restore()

	for i := 97; i < 123; i++ {
		for j := 97; j < 123; j++ {
			value := string([]byte{byte(i), byte(j)})
			key := "scan:" + value
			conn.Do("set", key, value)
		}
	}

	cursor, result, err := conn.DoScan("scan", 0, "match", "scan:*", "count", 5)
	assert.Nil(err)
	assert.True(cursor != 0)
	assert.True(result.Len() > 0)

	loopCursor := 0
	loopCount := 0
	valueCount := 0
	for {
		cursor, result, err := conn.DoScan("scan", loopCursor, "match", "scan:*", "count", 5)
		assert.Nil(err)

		loopCount += 1
		valueCount += result.Len()

		if cursor == 0 {
			break
		}
		loopCursor = cursor
	}
	assert.True(loopCount > 1)
	assert.Equal(valueCount, 26*26)
}

func TestHash(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	conn, restore := connectDatabase(assert)
	defer restore()

	e := map[string]string{
		"e1": "foo",
		"e2": "bar",
		"e3": "yadda",
	}
	ok, err := conn.DoOK("hmset", "hash", "a", "foo", "b", 2, "c", 3.3, "d", true, "e", e)
	assert.Nil(err)
	assert.True(ok)

	valueA, err := conn.DoString("hget", "hash", "a")
	assert.Nil(err)
	assert.Equal(valueA, "foo")

	hash, err := conn.DoHash("hgetall", "hash")
	assert.Nil(err)
	assert.Length(hash, 5)
	valueA, err = hash.String("a")
	assert.Nil(err)
	assert.Equal(valueA, "foo")
	valueB, err := hash.Int("b")
	assert.Nil(err)
	assert.Equal(valueB, 2)
	valueC, err := hash.Float64("c")
	assert.Nil(err)
	assert.Equal(valueC, 3.3)
	valueD, err := hash.Bool("d")
	assert.Nil(err)
	assert.True(valueD)
	valueE := hash.StringMap("e")
	assert.Equal(valueE, e)
}

func TestHScan(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	conn, restore := connectDatabase(assert)
	defer restore()

	for i := 97; i < 123; i++ {
		for j := 97; j < 123; j++ {
			value := string([]byte{byte(i), byte(j)})
			field := "field:" + value
			conn.Do("hset", "scan-hash", field, value)
		}
	}

	cursor, result, err := conn.DoScan("hscan", "scan-hash", 0, "match", "field:*", "count", 5)
	assert.Nil(err)
	assert.True(cursor != 0)
	assert.True(result.Len() > 0)

	loopCursor := 0
	loopCount := 0
	valueCount := 0
	for {
		cursor, result, err := conn.DoScan("hscan", "scan-hash", loopCursor, "match", "field:*", "count", 5)
		assert.Nil(err)
		hash, err := result.Hash()
		assert.Nil(err)

		loopCount += 1
		valueCount += hash.Len()

		if cursor == 0 {
			break
		}
		loopCursor = cursor
	}
	assert.True(loopCount > 1)
	assert.Equal(valueCount, 26*26)
}

func TestList(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	conn, restore := connectDatabase(assert)
	defer restore()

	pushed, err := conn.DoInt("lpush", "list", 1, 2, 3, 4, 5)
	assert.Nil(err)
	assert.Equal(pushed, 5)

	popped, err := conn.DoInt("lpop", "list")
	assert.Nil(err)
	assert.Equal(popped, 5)
}

func TestSet(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	conn, restore := connectDatabase(assert)
	defer restore()

	added, err := conn.DoInt("sadd", "set", 1, 2, 3, 4, 5)
	assert.Nil(err)
	assert.Equal(added, 5)

	is, err := conn.DoBool("sismember", "set", 2)
	assert.Nil(err)
	assert.True(is)
	is, err = conn.DoBool("sismember", "set", 99)
	assert.Nil(err)
	assert.False(is)

	rand, err := conn.DoInt("srandmember", "set")
	assert.Nil(err)
	assert.True(rand >= 1 && rand <= 5)
}

func TestSScan(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	conn, restore := connectDatabase(assert)
	defer restore()

	for i := 97; i < 123; i++ {
		for j := 97; j < 123; j++ {
			value := string([]byte{byte(i), byte(j)})
			conn.Do("sadd", "scan-set", value)
		}
	}

	cursor, result, err := conn.DoScan("sscan", "scan-set", 0, "match", "*", "count", 5)
	assert.Nil(err)
	assert.True(cursor != 0)
	assert.True(result.Len() > 0)

	loopCursor := 0
	loopCount := 0
	valueCount := 0
	for {
		cursor, result, err := conn.DoScan("sscan", "scan-set", loopCursor, "match", "*", "count", 5)
		assert.Nil(err)

		loopCount += 1
		valueCount += result.Len()

		if cursor == 0 {
			break
		}
		loopCursor = cursor
	}
	assert.True(loopCount > 1)
	assert.Equal(valueCount, 26*26)
}

func TestSortedSet(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	conn, restore := connectDatabase(assert)
	defer restore()

	added, err := conn.DoInt("zadd", "sorted-set", 1, "a", 2, "b", 3, "c", 4, "d", 5, "e")
	assert.Nil(err)
	assert.Equal(added, 5)

	scoredValues, err := conn.DoScoredValues("zrange", "sorted-set", 2, 4)
	assert.Nil(err)
	assert.Length(scoredValues, 3)
	assert.Equal(scoredValues[0].Score, 0.0)

	scoredValues, err = conn.DoScoredValues("zrange", "sorted-set", 2, 4, "withscores")
	assert.Nil(err)
	assert.Length(scoredValues, 3)
	assert.Equal(scoredValues[0].Score, 3.0)
	assert.Equal(scoredValues[1].Score, 4.0)
	assert.Equal(scoredValues[2].Score, 5.0)
}

func TestZScan(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	conn, restore := connectDatabase(assert)
	defer restore()

	for i := 97; i < 123; i++ {
		for j := 97; j < 123; j++ {
			score := i * j
			value := string([]byte{byte(i), byte(j)})
			conn.Do("zadd", "scan-sorted-set", score, value)
		}
	}

	cursor, result, err := conn.DoScan("zscan", "scan-sorted-set", 0, "match", "*", "count", 5)
	assert.Nil(err)
	assert.True(cursor != 0)
	assert.True(result.Len() > 0)

	loopCursor := 0
	loopCount := 0
	valueCount := 0
	for {
		cursor, result, err := conn.DoScan("zscan", "scan-sorted-set", loopCursor, "match", "*", "count", 5)
		assert.Nil(err)

		loopCount += 1
		scoredValues, err := result.ScoredValues(true)
		assert.Nil(err)
		valueCount += scoredValues.Len()

		if cursor == 0 {
			break
		}
		loopCursor = cursor
	}
	assert.True(loopCount > 1)
	assert.Equal(valueCount, 26*26)
}

func TestTransactionConnection(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	conn, restore := connectDatabase(assert)
	defer restore()

	ok, err := conn.DoOK("multi")
	assert.Nil(err)
	assert.True(ok)
	conn.Do("set", "tx:a", 1)
	conn.Do("set", "tx:b", 2)
	conn.Do("set", "tx:c", 3)
	result, err := conn.Do("exec")
	assert.Nil(err)
	assert.Length(result, 3)
	valueB, err := conn.DoInt("get", "tx:b")
	assert.Nil(err)
	assert.Equal(valueB, 2)

	ok, err = conn.DoOK("multi")
	assert.Nil(err)
	assert.True(ok)
	conn.Do("set", "tx:d", 4)
	conn.Do("set", "tx:e", 5)
	conn.Do("set", "tx:f", 6)
	ok, err = conn.DoOK("discard")
	assert.Nil(err)
	assert.True(ok)
	valueE, err := conn.DoValue("get", "tx:e")
	assert.Nil(err)
	assert.True(valueE.IsNil())

	sig := make(chan struct{})
	go func() {
		asyncConn, restore := connectDatabase(assert)
		defer restore()
		<-sig
		asyncConn.Do("set", "tx:h", 99)
		sig <- struct{}{}
	}()
	conn.Do("watch", "tx:h")
	ok, err = conn.DoOK("multi")
	assert.Nil(err)
	assert.True(ok)
	conn.Do("set", "tx:g", 4)
	conn.Do("set", "tx:h", 5)
	sig <- struct{}{}
	conn.Do("set", "tx:i", 6)
	<-sig
	_, err = conn.Do("exec")
	assert.True(errors.IsError(err, redis.ErrTimeout))
	valueH, err := conn.DoInt("get", "tx:h")
	assert.Nil(err)
	assert.Equal(valueH, 99)
}

func TestTransactionPipeline(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	conn, connRestore := connectDatabase(assert)
	defer connRestore()
	ppl, pplRestore := pipelineDatabase(assert)
	defer pplRestore()

	err := ppl.Do("multi")
	assert.Nil(err)
	ppl.Do("set", "pipeline:a", 1)
	ppl.Do("set", "pipeline:b", 2)
	ppl.Do("set", "pipeline:c", 3)
	ppl.Do("exec")
	results, err := ppl.Collect()
	assert.Nil(err)
	assert.Length(results, 5)
	valueB, err := conn.DoInt("get", "pipeline:b")
	assert.Nil(err)
	assert.Equal(valueB, 2)

	err = ppl.Do("multi")
	assert.Nil(err)
	ppl.Do("set", "pipeline:d", 4)
	ppl.Do("set", "pipeline:e", 5)
	ppl.Do("set", "pipeline:f", 6)
	ppl.Do("discard")
	results, err = ppl.Collect()
	assert.Nil(err)
	assert.Length(results, 5)
	valueE, err := conn.DoValue("get", "pipeline:e")
	assert.Nil(err)
	assert.True(valueE.IsNil())
}

func TestTransactionPipelineWatch(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	sigC := audit.MakeSigChan()
	// Background tasks.
	bgConn, bgConnRestore := connectDatabase(assert)
	defer bgConnRestore()
	go func() {
		<-sigC
		bgConn.Do("set", "watch:b", 99)
		sigC <- struct{}{}
	}()
	// Foreground tasks.
	fgConn, fgConnRestore := pipelineDatabase(assert)
	defer fgConnRestore()
	fgConn.Do("set", "watch:b", 0)
	fgConn.Do("watch", "watch:b")
	err := fgConn.Do("multi")
	assert.Nil(err)
	fgConn.Do("set", "watch:a", 1)
	fgConn.Do("set", "watch:b", 2)
	sigC <- struct{}{}
	fgConn.Do("set", "watch:c", 3)
	<-sigC
	fgConn.Do("exec")
	_, err = fgConn.Collect()
	assert.True(errors.IsError(err, redis.ErrTimeout))
	valueB, err := bgConn.DoInt("get", "watch:b")
	assert.Nil(err)
	assert.Equal(valueB, 99)
}

func TestScripting(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	conn, restore := connectDatabase(assert)
	defer restore()

	script := "return {KEYS[1],KEYS[2],ARGV[1],ARGV[2]}"
	result, err := conn.Do("eval", script, 2, "key1", "key2", 1, "two")
	assert.Nil(err)
	assert.Length(result, 4)
	key1, err := result.StringAt(0)
	assert.Nil(err)
	assert.Equal(key1, "key1")
	key2, err := result.StringAt(1)
	assert.Nil(err)
	assert.Equal(key2, "key2")
	argv1, err := result.IntAt(2)
	assert.Nil(err)
	assert.Equal(argv1, 1)
	argv2, err := result.StringAt(3)
	assert.Nil(err)
	assert.Equal(argv2, "two")

	script = "return {redis.error_reply('x'), 'x', redis.status_reply('x')}"
	result, err = conn.Do("eval", script, 0)
	assert.Nil(err)
	assert.Length(result, 3)
	reply1, err := result.StringAt(0)
	assert.Nil(err)
	assert.Equal(reply1, "-x")
	reply2, err := result.StringAt(1)
	assert.Nil(err)
	assert.Equal(reply2, "x")
	reply3, err := result.StringAt(2)
	assert.Nil(err)
	assert.Equal(reply3, "+x")
}

func TestPubSub(t *testing.T) {
	assert := audit.NewTestingAssertion(t, true)
	conn, connRestore := connectDatabase(assert)
	defer connRestore()
	sub, subRestore := subscribeDatabase(assert)
	defer subRestore()

	_, err := conn.Do("subscribe", "pubsub")
	assert.True(errors.IsError(err, redis.ErrUseSubscription))

	err = sub.Subscribe("pubsub")
	assert.Nil(err)
	pv, err := sub.Pop()
	assert.Nil(err)
	assert.Equal(pv.Kind, "subscribe")
	assert.Equal(pv.Channel, "pubsub")
	assert.Equal(pv.Count, 1)

	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(50 * time.Millisecond)
			receivers, err := conn.DoInt("publish", "pubsub", i)
			assert.Nil(err)
			assert.Equal(receivers, 1)
		}
	}()

	sleep := 1 * time.Millisecond
	for i := 0; i < 10; i++ {
		time.Sleep(sleep)
		pv, err := sub.Pop()
		assert.Nil(err)
		assert.Equal(pv.Kind, "message")
		assert.Equal(pv.Channel, "pubsub")
		value, err := pv.Value.Int()
		assert.Nil(err)
		assert.Equal(value, i)
		sleep *= 2
	}
}

// EOF
