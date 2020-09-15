package types

type ProblemsGORM struct {
	ProblemId int64  `gorm:"column:id"`
	UUID      string `gorm:"column:uuid"`
}

type TestcaseGORM struct {
	TestcaseID int64  `gorm:"column:id"`
	Name       string `gorm:"column:name"`
	Input      []byte
	Output     []byte
}
