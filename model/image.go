package model

type Image struct {
	Id          string `json:"id,omitempty" bson:"id,omitempty" form:"id,omitempty"`
	UserId      string `json:"user_id,omitempty" bson:"user_id,omitempty" form:"user_id,omitempty"`
	RequestId   string `json:"request_id,omitempty" bson:"request_id,omitempty" form:"request_id,omitempty"`
	Path        string `json:"path,omitempty" bson:"path,omitempty" form:"path,omitempty"`
	Data        string `json:"data,omitempty" bson:"data,omitempty" form:"data,omitempty"`
	DataType    string `json:"data_type,omitempty" bson:"data_type,omitempty" form:"data_type,omitempty"`
	StorageType string `json:"storage_type,omitempty" bson:"storage_type,omitempty" form:"storage_type,omitempty"`
	Status      string `json:"status,omitempty" bson:"status,omitempty" form:"status,omitempty"`
	Error       string `json:"error,omitempty" bson:"error,omitempty" form:"error,omitempty"`
	CreateAt    string `json:"create_at,omitempty" bson:"create_at,omitempty" form:"create_at,omitempty"`
}
