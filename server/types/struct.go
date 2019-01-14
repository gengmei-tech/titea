package types

// Namespace avoid package cycle include
type Namespace struct {
	// 名称 group.service
	Name string `json:"name"`

	Type uint8 `json:"type"`
	// 命名空间所在的组
	Group string `json:"group"`

	// 名空间所在的service
	Service string `json:"service"`

	// 创建人
	Creator string `json:"creator"`

	// 创建时间
	CreateAt int64 `json:"createAt"`

	// 对应数据库的索引
	Index uint64 `json:"index"`
}
