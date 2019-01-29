package namespace

import (
	"encoding/json"
	"github.com/gengmei-tech/titea/server/store"
	"github.com/gengmei-tech/titea/server/terror"
	"github.com/gengmei-tech/titea/server/types"
	"regexp"
	"strings"
	"time"
)

// namespace in tikv is store in key => value format

// Namespace type
const (
	// TYPE_SYSTEM system namespace
	TYPESYSTEM uint8 = 1
	// TYPE_DEFAULT default namespace
	TYPEDEFAULT uint8 = 2
	// TYPE_CUSTOM custom namespace
	TYPECUSTOM uint8 = 3
)

var (
	defaultSeparator = "."
	defaultNamespace = "default.default"
	systemNamespace  = "system.system"
	systemCreator    = "system"

	// namespace key kvsystem.system|m|namespace
	namespaceKey = []byte(types.SystemPrefix + "system.systemmnamespace")

	// name => namespace  json store in tikv
	namespaces = make(map[string]*types.Namespace)

	// dbindex => namespace
	dbs = make(map[uint64]*types.Namespace)

	// namespace regex
	reg = regexp.MustCompile(`^[\w]+\.[\w]+$`)
)

// SelectNamespace implement select db command
func SelectNamespace(dbindex uint64) (*types.Namespace, error) {
	if namespace, ok := dbs[dbindex]; ok {
		return namespace, nil
	}
	return nil, terror.ErrNsNotExist
}

// RegisterNamespace register namespace, creator for record
func RegisterNamespace(db *store.Store, name, creator string, dbindex uint64) error {
	// check group.service valid
	if !checkValidNs(name) || dbindex == 0 {
		return terror.ErrCmdParams
	}

	// check exists
	if _, ok := namespaces[name]; ok {
		return terror.ErrNsExist
	}
	if _, ok := dbs[dbindex]; ok {
		return terror.ErrNsExist
	}
	namespace, err := registerNamespace(db, name, creator, dbindex, TYPECUSTOM)
	if err := db.Commit(); err != nil {
		return terror.ErrBackendType
	}
	// add to expire and gc
	store.SyncExpireGc(db, &namespace)
	return err
}

// LoadNamespace load namespace from tikv after server start
func LoadNamespace(db *store.Store) {
	value, err := db.Get(namespaceKey)
	if err != nil {
		panic("LoadNamespace Error")
	}
	if value != nil {
		if err := json.Unmarshal(value, &namespaces); err != nil {
			panic("json.Unmarshal AllNamespace Error")
		}
		for _, namespace := range namespaces {
			if namespace.Type != TYPESYSTEM {
				dbs[namespace.Index] = namespace
			}
		}
	} else {
		// default
		initNamespace(db)
	}
	return
}

// GetDefaultNamespace default namespace
func GetDefaultNamespace() *types.Namespace {
	namespace, ok := namespaces[defaultNamespace]
	if ok {
		return namespace
	}
	return nil
}

// GetAllNamespace get all
func GetAllNamespace() map[string]*types.Namespace {
	return namespaces
}

// check group.service is valid
func checkValidNs(name string) bool {
	if !reg.MatchString(name) {
		return false
	}
	// "default" invalid
	if strings.Contains(name, "default") {
		return false
	}
	// "system" invalid
	if strings.Contains(name, "system") {
		return false
	}
	return true
}

func registerNamespace(db *store.Store, name, creator string, dbindex uint64, kind uint8) (types.Namespace, error) {
	tmp := strings.Split(name, defaultSeparator)
	namespace := types.Namespace{
		Name:     name,
		Group:    tmp[0],
		Service:  tmp[1],
		Creator:  creator,
		Index:    dbindex,
		CreateAt: time.Now().Unix(),
		Type:     kind,
	}
	namespaces[name] = &namespace
	dbs[dbindex] = &namespace
	err := saveNamespace(db)
	return namespace, err
}

func initNamespace(db *store.Store) {
	db.WriteReset()
	if namespace, err := registerNamespace(db, defaultNamespace, systemCreator, 0, TYPEDEFAULT); err != nil {
		panic("Init Namespace Error")
	} else {
		dbs[namespace.Index] = &namespace
	}

	// Init system namespace
	if _, err := registerNamespace(db, systemNamespace, systemCreator, 0, TYPESYSTEM); err != nil {
		panic("Init Namespace Error")
	}
	if err := db.Commit(); err != nil {
		panic("Init Namespace Failed")
	}
}

func saveNamespace(db *store.Store) error {
	value, _ := json.Marshal(&namespaces)
	return db.Set(namespaceKey, value)
}

// DeleteNamespace group.service
// persist delete in TiKv
func DeleteNamespace(db *store.Store, name string, persist bool) error {
	// check valid
	if !checkValidNs(name) {
		return terror.ErrCmdParams
	}

	// check exists
	if _, ok := namespaces[name]; !ok {
		return terror.ErrNsNotExist
	}
	dbindex := namespaces[name].Index
	delete(namespaces, name)
	if persist {
		saveNamespace(db)
	}
	if dbindex > 0 {
		delete(dbs, dbindex)
	}
	return nil
}
