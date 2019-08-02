package rados_test

import (
	"testing"
	"fmt"
	"sync"

	"github.com/ceph/go-ceph/rados"
	"github.com/stretchr/testify/assert"
)

func TestAsyncWrite(t *testing.T) {
	conn, _ := rados.NewConn()
	conn.ReadDefaultConfigFile()
	conn.Connect()

	// make pool
	/*
	pool_name := GetUUID()
	err := conn.MakePool(pool_name)
	assert.NoError(t, err)
	*/
	pool_name := "y0"

	pool, err := conn.OpenIOContext(pool_name)
	assert.NoError(t, err)

	completion, err := rados.NewCompletion()
	assert.NoError(t, err)

	defer completion.Release()

	isComplete := completion.IsComplete()
	assert.Equal(t, isComplete, false)

	isSafe := completion.IsSafe()
	assert.Equal(t, isSafe, false)

	bytes_in := []byte("input data")

	err = pool.AsyncWrite("obj", bytes_in, 0, completion)
	assert.NoError(t, err)

	completion.WaitForComplete()
	completion.WaitForSafe()

	isComplete = completion.IsComplete()
	assert.Equal(t, isComplete, true)

	isSafe = completion.IsSafe()
	assert.Equal(t, isSafe, true)

	ret := completion.ReturnValue()
	assert.Equal(t, ret, 0)

	pool.Destroy()
	conn.Shutdown()
}

func TestMultiAsyncDelete(t *testing.T) {
	conn, _ := rados.NewConn()
	conn.ReadDefaultConfigFile()
	conn.Connect()

	// make pool
	/*
	pool_name := GetUUID()
	err := conn.MakePool(pool_name)
	assert.NoError(t, err)
	*/
	pool_name := "y0"

	pool, err := conn.OpenIOContext(pool_name)
	assert.NoError(t, err)

	objects := 100
	for i := 0; i < objects; i++ {
		key := fmt.Sprintf("obj-%d", i)
		bytes_in := []byte(fmt.Sprintf("input data%d", i))
		err = pool.Write(key, bytes_in, 0)
		assert.NoError(t, err)
	}

	var completions []*rados.Completion
	for i := 0; i < objects; i++ {
		completion, err := rados.NewCompletion()
		assert.NoError(t, err)

		key := fmt.Sprintf("obj-%d", i)
		//err = pool.AsyncDelete(completion, key)
		err = pool.AsyncWrite(key, []byte("bbbbbbb"), 0, completion)
		assert.NoError(t, err)

		completions = append(completions, completion)
	}

	var wg sync.WaitGroup
	for _, completion := range completions {
		wg.Add(1)
		go func(completion *rados.Completion) {
			defer wg.Done()

			completion.WaitForComplete()
			fmt.Println("Done completion")
			completion.WaitForSafe()
			fmt.Println("Done waitforsafe")

			isComplete := completion.IsComplete()
			assert.Equal(t, isComplete, true)

			isSafe := completion.IsSafe()
			assert.Equal(t, isSafe, true)

			ret := completion.ReturnValue()
			assert.Equal(t, ret, 0)
		}(completion)
	}
	wg.Wait()

	for i :=0; i < objects; i++ {
		 key := fmt.Sprintf("obj-%d", i)
		 var buf []byte
		 buf = make([]byte, 15)
		l, err := pool.Read(key, buf, 0)	
		fmt.Printf("key %s val %s len %d err %v\n", key, buf, l, err)
	 }

	pool.Destroy()
	conn.Shutdown()
}
