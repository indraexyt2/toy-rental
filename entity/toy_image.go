package entity

type ToyImage struct {
	BaseEntity
	ImageURL string `gorm:"type:text;not null" json:"image_url"`

	Toy []Toy `gorm:"many2many:toy_toy_images" json:"-"`
}

func (*ToyImage) TableName() string {
	return "toy_images"
}
