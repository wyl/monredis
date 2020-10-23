package monredismap

import (
	"github.com/garyburd/redigo/redis"
)

// RedisWriter is an interface for types that can write to redis using send/flush (pipelined operations)
// The main purpose of breaking this out into an interface is for ease of mocking in tests.
type RedisWriter interface {
	Send(cmd string, args ...interface{}) error
	Flush() error
}

type redisWriter struct {
	conn          redis.Conn
	flushInterval int
	currentCount  int
}

// NewRedisWriter creates a new RedisWriter.  We wrap redis.Conn here so that we can specify how many
// documents we want to allow buffered before flushing automatically.
func NewRedisWriter(conn redis.Conn) RedisWriter {
	writer := &redisWriter{
		conn:          conn,
		flushInterval: 100,
	}
	return writer
}

// Send uses the same interface as redis.Conn.Send().  The difference is that RedisWriter's Send
// method takes care of automatically flushing after flushInterval amount of documents.
func (r *redisWriter) Send(cmd string, args ...interface{}) error {
	if err := r.conn.Send(cmd, args...); err != nil {
		return err
	}
	r.currentCount++
	if r.currentCount >= r.flushInterval {
		if err := r.Flush(); err != nil {
			return err
		}
		r.currentCount = 0
		// Do a ping and wait for the reply to ensure that there is no data waiting to be received.
		if _, err := r.conn.Do("PING"); err != nil {
			return err
		}
	}
	return nil

}

// Flush triggers a flush on the underlying redis connection.
func (r *redisWriter) Flush() error {
	return r.conn.Flush()
}
