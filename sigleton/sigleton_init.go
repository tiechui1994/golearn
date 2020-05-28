package sigleton

type sigleton struct {
}

var instance *sigleton

func init() {
	instance = &sigleton{}
}

func GetInstance() *sigleton {
	return instance
}
