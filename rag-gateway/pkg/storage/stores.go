package storage

import (
	"context"
	"time"
)

type WatchEventType string

const (
	// An operation that creates a new key OR modifies an existing key.
	//
	// NB: The Watch API does not distinguish between create and modify events.
	// It is not practical (nor desired, in most cases) to provide this info
	// to the caller, because it cannot be guaranteed to be accurate in all cases.
	// Because of the inability to make this guarantee, any client code that
	// relies on this distinction would be highly likely to end up in an invalid
	// state after a sufficient amount of time, or after issuing a watch request
	// on a key that has a complex and/or truncated history. However, in certain
	// cases, clients may be able to correlate events with out-of-band information
	// to reliably disambiguate Put events. This is necessarily an implementation
	// detail and may not always be possible.
	WatchEventPut WatchEventType = "Put"

	// An operation that removes an existing key.
	//
	// Delete events make few guarantees, as different backends handle deletes
	// differently. Backends are not required to discard revision history, or
	// to stop sending events for a key after it has been deleted. Keys may
	// be recreated after a delete event, in which case a Put event will follow.
	// Such events may or may not contain a previous revision value, depending
	// on implementation details of the backend (they will always contain a
	// current revision value, though).
	WatchEventDelete WatchEventType = "Delete"
)

type WatchEvent[T any] struct {
	EventType WatchEventType
	Current   T
	Previous  T
}

type KeyValueStoreT[T any] interface {
	Put(ctx context.Context, key string, value T, opts ...PutOpt) error
	Get(ctx context.Context, key string, opts ...GetOpt) (T, error)

	// Starts a watch on the specified key. The returned channel will receive
	// events for the key until the context is canceled, after which the
	// channel will be closed. This function does not block. An error will only
	// be returned if the key is invalid or the watch fails to start.
	//
	// When the watch is started, the current value of the key will be sent
	// if and only if both of the following conditions are met:
	// 1. A revision is explicitly set in the watch options. If no revision is
	//    specified, only future events will be sent.
	// 2. The key exists; or in prefix mode, there is at least one key matching
	//    the prefix.
	//
	// In most cases a starting revision should be specified, as this will
	// ensure no events are missed.
	//
	// This function can be called multiple times for the same key, prefix, or
	// overlapping prefixes. Each call will initiate a separate watch, and events
	// are always replicated to all active watches.
	//
	// The channels are buffered to hold at least 64 events. Ensure that events
	// are read from the channel in a timely manner if a large volume of events
	// are expected; otherwise it will block and events may be delayed, or be
	// dropped by the backend.
	Watch(ctx context.Context, key string, opts ...WatchOpt) (<-chan WatchEvent[KeyRevision[T]], error)
	Delete(ctx context.Context, key string, opts ...DeleteOpt) error
	ListKeys(ctx context.Context, prefix string, opts ...ListOpt) ([]string, error)
	History(ctx context.Context, key string, opts ...HistoryOpt) ([]KeyRevision[T], error)
}

type KeyValueStore = KeyValueStoreT[[]byte]

type KeyRevision[T any] interface {
	Key() string
	SetKey(string)

	// If values were requested, returns the value at this revision. Otherwise,
	// returns the zero value for T.
	// Note that if the value has a revision field, it will *not*
	// be populated, and should be set manually if needed using the Revision()
	// method.
	Value() T
	// Returns the revision of this key. Larger values are newer, but the
	// revision number should otherwise be treated as an opaque value.
	Revision() int64
	// Returns the timestamp of this revision. This may or may not always be
	// available, depending on if the underlying store supports it.
	Timestamp() time.Time
}

type KeyRevisionImpl[T any] struct {
	K    string
	V    T
	Rev  int64
	Time time.Time
}

func (k *KeyRevisionImpl[T]) Key() string {
	return k.K
}

func (k *KeyRevisionImpl[T]) SetKey(key string) {
	k.K = key
}

func (k *KeyRevisionImpl[T]) Value() T {
	return k.V
}

func (k *KeyRevisionImpl[T]) Revision() int64 {
	return k.Rev
}

func (k *KeyRevisionImpl[T]) Timestamp() time.Time {
	return k.Time
}

type KeyValueStoreTBroker[T any] interface {
	KeyValueStore(namespace string) KeyValueStoreT[T]
}

type ValueStoreT[T any] interface {
	Put(ctx context.Context, value T, opts ...PutOpt) error
	Get(ctx context.Context, opts ...GetOpt) (T, error)
	Watch(ctx context.Context, opts ...WatchOpt) (<-chan WatchEvent[KeyRevision[T]], error)
	Delete(ctx context.Context, opts ...DeleteOpt) error
	History(ctx context.Context, opts ...HistoryOpt) ([]KeyRevision[T], error)
}
