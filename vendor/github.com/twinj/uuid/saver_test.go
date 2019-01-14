package uuid

import (
	"errors"
	"gopkg.in/stretchr/testify.v1/assert"
	"sync"
	"testing"
	"time"
)

type save struct {
	saved bool
	store *Store
	err   error
	sync.Mutex
}

func (o *save) Init() Saver {
	return o
}

func (o *save) Save(pStore Store) {
	o.Lock()
	defer o.Unlock()
	o.saved = true
}

func (o *save) Read() (Store, error) {
	if o.store != nil {
		return *o.store, nil
	}
	if o.err != nil {
		return Store{}, o.err
	}
	return Store{}, nil
}

func TestRegisterSaver(t *testing.T) {
	registerTestGenerator(Timestamp(2048), []byte{0xaa})

	saver := &save{store: &Store{}}
	RegisterSaver(saver)

	assert.NotNil(t, generator.Saver, "Saver should save")

	assert.Panics(t, func() {
		RegisterSaver(saver)
	})

	registerDefaultGenerator()
}

func TestSaverRead(t *testing.T) {
	now, node := registerTestGenerator(Now().Sub(time.Second), []byte{0xaa})

	storageStamp, err := registerSaver(now.Sub(time.Second*2), node)

	assert.NoError(t, err)
	assert.NotNil(t, generator.Saver, "Saver should save")
	assert.NotNil(t, generator.Store, "Default generator store should not return an empty store")
	assert.Equal(t, Sequence(2), generator.Store.Sequence, "Successfull read should have actual given sequence")
	assert.True(t, generator.Store.Timestamp > storageStamp, "Failed read should generate a time")
	assert.NotEmpty(t, generator.Store.Node, "There should be a node id")

	// Read returns an error
	_, node = registerTestGenerator(Now(), []byte{0xaa})
	saver := &save{err: errors.New("Read broken")}
	RegisterSaver(saver)

	assert.Nil(t, generator.Saver, "Saver should not exist")
	assert.NotNil(t, generator.Store, "Default generator store should not return an empty store")
	assert.NotEqual(t, Sequence(0), generator.Sequence, "Failed read should generate a non zero random sequence")
	assert.True(t, generator.Timestamp > 0, "Failed read should generate a time")
	assert.Equal(t, node, generator.Node, "There should be a node id")
	registerDefaultGenerator()
}

func TestSaverSave(t *testing.T) {
	registerTestGenerator(Now().Add(1024), nodeBytes)

	saver := &save{}
	RegisterSaver(saver)

	NewV1()

	saver.Lock()
	defer saver.Unlock()

	assert.True(t, saver.saved, "Saver should save")
	registerDefaultGenerator()
}

func TestStore_String(t *testing.T) {
	store := &Store{Node: []byte{0xdd, 0xee, 0xff, 0xaa, 0xbb}, Sequence: 2, Timestamp: 3}
	assert.Equal(t, "Timestamp[2167-05-04 23:34:33.709551916 +0000 UTC]-Sequence[2]-Node[ddeeffaabb]", store.String(), "The output store string should match")
}

func registerSaver(storageStamp Timestamp, pNode Node) (ss Timestamp, err error) {
	ss = storageStamp

	once = new(sync.Once)
	saver := &save{store: &Store{Node: pNode, Sequence: 2, Timestamp: storageStamp}}
	err = RegisterSaver(saver)
	return
}
