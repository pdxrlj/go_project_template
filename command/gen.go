package main

import (
	"telecommunications_repair_hub/config"
	"telecommunications_repair_hub/models"
	"telecommunications_repair_hub/models/query"
	"telecommunications_repair_hub/pkg/db"

	"gorm.io/gen"
)

func main() {
	cfg := config.InitConfig()

	g := gen.NewGenerator(gen.Config{
		OutPath:           "./models/query",
		ModelPkgPath:      "../models",
		Mode:              gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
		FieldNullable:     true,
		FieldSignable:     true,
		WithUnitTest:      true,
		FieldCoverable:    true,
		FieldWithIndexTag: true,
		FieldWithTypeTag:  true,
	})

	db, err := db.New(cfg)
	if err != nil {
		panic(err)
	}

	query.SetDefault(db.DB)

	g.UseDB(db.DB)
	g.ApplyBasic(
		&models.User{},
	)
	g.Execute()
}
